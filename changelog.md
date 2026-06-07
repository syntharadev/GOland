# 📝 Changelog / Historial de Versiones

## [1.0.0] - v1.0.0 - Hackathon Edition (2026-06-07)

### 🚀 Añadido
*   **Integración Multicloud Completa**: Configuración de backend modular en Go preparado para producción desplegado en **Google Cloud Run**, base de datos principal en **MongoDB Atlas** y autenticación de usuarios mediante **Supabase (Google OAuth2)**.
*   **Agentic UI & Swarm Router**: 
    *   Enrutamiento semántico automático de intenciones (*ClasificarIntencion* con Gemini) hacia uno de los 4 escuadrones de GOnions.
    *   Soporte dinámico para los 9 agentes (GOnions), cada uno provisto de su respectivo prompt de `System Instruction` y comportamiento de personalidad único.
*   **Ajuste Térmico Diferenciado (Temperature)**:
    *   Asignación de temperaturas del modelo por el tipo de rol de agente de IA:
        *   **0.1 - 0.2** para agentes analíticos y precisos (*La Guardiana*, *La Ingeniera*, *La Bibliotecaria*, *El Mensajero*, *El Senior*, *La Cronometradora*).
        *   **0.5 - 0.6** para agentes de guía didáctica (*El Profesor*, *El Constructor*).
        *   **1.0** para el agente caótico y experimental (*El Hacker*).
*   **Memoria Conversacional del Swarm**:
    *   Uso de `StartChat()` para mantener el flujo de diálogo e historial interactivo.
    *   Aislamiento seguro de sesiones en memoria con `sync.Map` mapeadas a la combinación de `UserNick` y `AgentName`.
*   **Inyección Silenciosa de Contexto**:
    *   Inyección automática del código del editor activo de forma oculta en el backend antes de realizar la petición al modelo.
*   **Frontend Premium**:
    *   Refactorización del estado activo de las cartas GOnion en `app_GOland.html` para transición fluida de texto debajo de la imagen al hacer click con opacidad `1.0` completa y glassmorphism.
*   **Licencia**:
    *   Adición formal de los términos de la Licencia MIT.

### 🛠️ Corrección de Errores y Seguridad
*   **OAuth Env Injections**: Inyectadas las credenciales de Google OAuth directamente en la revisión de Cloud Run, corrigiendo el error HTTP 400.
*   **Supabase RLS**: Remediación de seguridad al eliminar la tabla huérfana y sin protección RLS `progreso_usuarios` en Supabase mediante drop en cascada.
*   **RAG robusto**: Reemplazo de embeddings remotos de Gemini por un sistema local y veloz RAG MongoDB basado en Regex que elude el bloqueo geográfico de la API de embeddings de Google.
