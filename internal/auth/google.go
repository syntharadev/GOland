package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/gorilla/sessions"
)

// Configuración de OAuth2
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// Almacén de cookies encriptadas
var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))

func generateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, HttpOnly: true, Path: "/"}
	http.SetCookie(w, &cookie)
	return state
}

// HandleGoogleLogin redirige al usuario a la pantalla de Google
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthStateString := generateStateOauthCookie(w)
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback procesa la respuesta de Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// 1. Validar estado (CSRF protection)
	oauthState, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthState.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// 2. Intercambiar código por Token
	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// 3. Obtener datos del usuario
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer response.Body.Close()

	var userInfo struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	json.NewDecoder(response.Body).Decode(&userInfo)

	// 4. Crear sesión
	session, _ := Store.Get(r, "goland-session")
	session.Values["user_nick"] = userInfo.Name // Usamos el nombre real de Google
	session.Values["authenticated"] = true
	session.Save(r, w)

	// OJO: Aquí deberías llamar a db.SaveProgress si es un usuario nuevo (Nivel 1)

	// Redirigir al panel principal
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleAuthStatus devuelve si el usuario está autenticado y su nick
func HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "goland-session")
	
	nick, ok := session.Values["user_nick"].(string)
	authenticated, _ := session.Values["authenticated"].(bool)
	
	w.Header().Set("Content-Type", "application/json")
	if ok && authenticated {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": true,
			"nick": nick,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
	}
}
