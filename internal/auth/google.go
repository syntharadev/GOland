package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Configuración de OAuth2
var googleOauthConfig = &oauth2.Config{
	Scopes:   []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: google.Endpoint,
}

// Almacén de cookies encriptadas (se inicializa formalmente en Init())
var Store = sessions.NewCookieStore([]byte("secret-key-fallback-for-compilation-init-2026"))

// Init inicializa las variables de entorno de autenticación y configura la cookie de sesión de forma segura
func Init() {
	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	secret := os.Getenv("SESSION_SECRET_KEY")
	if secret == "" {
		secret = "goland_super_clave_secreta_en_produccion_2026"
	}
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 días
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

// RequireAuth es un middleware para proteger endpoints de la API
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "goland-session")
		authVal, ok := session.Values["authenticated"].(bool)
		if !ok || !authVal {
			http.Error(w, "No autorizado. Inicie sesión primero.", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	return state
}

// HandleGoogleLogin redirige al usuario a la pantalla de Google
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if redirectURL := os.Getenv("REDIRECT_URL"); redirectURL != "" {
		googleOauthConfig.RedirectURL = redirectURL
	} else {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		googleOauthConfig.RedirectURL = scheme + "://" + r.Host + "/auth/google/callback"
	}

	oauthStateString := generateStateOauthCookie(w)
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback procesa la respuesta de Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if redirectURL := os.Getenv("REDIRECT_URL"); redirectURL != "" {
		googleOauthConfig.RedirectURL = redirectURL
	} else {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		googleOauthConfig.RedirectURL = scheme + "://" + r.Host + "/auth/google/callback"
	}

	// 1. Validar estado (CSRF protection)
	oauthState, _ := r.Cookie("oauthstate")
	if oauthState == nil || r.FormValue("state") != oauthState.Value {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// 2. Intercambiar código por Token
	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// 3. Obtener datos del usuario
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
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
	http.Redirect(w, r, "/workspace", http.StatusSeeOther)
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
			"nick":          nick,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
	}
}

// HandleLogout limpia la sesión cuántica y redirige al index
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "goland-session")
	session.Values["authenticated"] = false
	delete(session.Values, "user_nick")
	session.Options.MaxAge = -1 // Expirar cookie inmediatamente
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
