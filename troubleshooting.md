# 🛠️ GOland - Troubleshooting & Incidencias Resueltas

Este documento registra los incidentes técnicos de infraestructura, seguridad y lógica LLM identificados y solventados durante la preparación para producción.

---

### 🚨 Incidente 1: Error 400 `invalid_request` (Missing `client_id`) en Login OAuth
*   **Síntoma:** Al intentar autenticarse en el frontend con el botón de Google Login, la API devolvía un error de mala petición indicando que el parámetro `client_id` no estaba definido.
*   **Causa Raíz:** Las credenciales de Google OAuth estaban configuradas en la interfaz de Supabase, pero el backend escrito en Go ejecutándose en Google Cloud Run no tenía inyectadas las variables de entorno para construir de forma segura la petición OAuth2 en los handlers de inicio.
*   **Solución:** Se configuró la inyección directa de las variables de entorno `GOOGLE_CLIENT_ID` y `GOOGLE_CLIENT_SECRET` en la revisión de despliegue y variables de entorno del contenedor en **Google Cloud Run**, asegurando que el backend las lea correctamente con `os.Getenv` en tiempo de ejecución.

---

### 🚨 Incidente 2: Alarma de Seguridad `rls_disabled_in_public` en Supabase
*   **Síntoma:** Auditoría de seguridad de Supabase marcaba en rojo la vulnerabilidad debido al bypass de políticas de seguridad.
*   **Causa Raíz:** Presencia de una tabla huérfana llamada `progreso_usuarios` en el esquema público de PostgreSQL sin políticas activadas de RLS (*Row Level Security*), lo que permitía posibles consultas no autorizadas si alguien interceptaba la URL de la base de datos pública.
*   **Solución:** Reducción de la superficie de ataque. Dado que la persistencia principal del estado del juego y los logs del enjambre residen formalmente en MongoDB Atlas, se eliminó la redundancia ejecutando `DROP TABLE public.progreso_usuarios CASCADE;` en la consola SQL de Supabase, remediando por completo la vulnerabilidad.

---

### 🚨 Incidente 3: Amnesia Conversacional y Falta de Contexto en los Agentes (GOnions)
*   **Síntoma:** Los agentes respondían de manera aislada (*stateless*). Olvidaban la consulta o explicación anterior y carecían de conocimiento sobre el código que el estudiante tenía cargado en el editor interactivo de la pantalla.
*   **Causa Raíz:** El router de swarms ejecutaba llamadas independientes a la API de Gemini usando `GenerateContent` puro de forma aislada, sin pasar historial y sin adjuntar el contexto del editor de código al prompt final.
*   **Solución:** Refactorización integral del router en `internal/router/swarms.go`:
    1. Se migró de `GenerateContent` a la API de chat del SDK de Go utilizando `StartChat()`.
    2. Se introdujo un almacén de sesiones seguro concurrente (`sync.Map`) indexado por la clave combinada `UserNick_AgentName` para retener las instancias de chat individualizadas de manera segura por usuario.
    3. Se implementó una inyección oculta de contexto que extrae el código del payload del frontend y lo concatena al final del mensaje de forma invisible para el usuario:
       `"Mensaje del usuario: {mensaje}. Contexto del código actual: {código}"`.
