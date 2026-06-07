# 🚀 GOland - Agentic UI Ecosystem (v1.0)

## Resumen Ejecutivo
GOland es una plataforma EdTech gamificada que revoluciona el aprendizaje del lenguaje Go. Mediante una arquitectura "Agentic UI", el usuario interactúa con un enjambre de 9 agentes de IA especializados (GOnions) que evalúan, guían y retan al estudiante en tiempo real.

## Arquitectura de Sistemas
*   **Frontend:** Vanilla JavaScript, HTML5, CSS3 (CSS Grid, Flexbox, Glassmorphism). Interfaz reactiva sin frameworks pesados.
*   **Backend / Orquestador:** Golang (`cmd/server/main.go`). Gestiona concurrencia, enrutamiento de IA y peticiones HTTP.
*   **Motor de Inteligencia Artificial:** Google Gemini API (SDK de Go). Uso avanzado de `System Instructions`, control de `Temperature` por perfil y sesiones de chat concurrentes.
*   **Autenticación:** Supabase (OAuth 2.0 con Google y GitHub).
*   **Base de Datos Principal:** MongoDB (Almacena progreso de usuarios, niveles y logs de sesión).
*   **Infraestructura Cloud:** Google Cloud Run (Contenedores Docker serverless, escalado automático).

## El Enjambre: Los 9 GOnions
1.  **El Profesor (Temp 0.6):** Guía socrático. Enseña paso a paso.
2.  **La Guardiana (Temp 0.1):** Implacable. Busca edge cases y vulnerabilidades.
3.  **La Bibliotecaria (Temp 0.2):** Extrae documentación oficial exacta de Go.
4.  **La Ingeniera (Temp 0.2):** Especialista en concurrencia (Goroutines, Channels).
5.  **El Hacker (Temp 1.0):** Respuestas poco convencionales, trucos y one-liners.
6.  **El Senior (Temp 0.3):** Revisor estricto de código limpio y filosofía *Effective Go*.
7.  **El Constructor (Temp 0.5):** Herramientas CI/CD y DevOps (Tool calling integrado con GitLab).
8.  **La Cronometradora (Temp 0.2):** Análisis de rendimiento y notación Big O (Tool calling con Elastic).
9.  **El Mensajero (Temp 0.3):** Gestor de logs de sistema y eventos de red.

## Flujo de Datos
1.  Frontend envía JSON: `{ "agent": "Profesor", "message": "...", "code": "..." }`.
2.  Backend intercepta, localiza o crea la sesión en `sync.Map` por usuario/agente.
3.  Asigna el `System Instruction` y `Temperature` correspondiente.
4.  Inyecta el código de forma invisible en el prompt.
5.  Gemini procesa y devuelve el stream/respuesta.
6.  Backend guarda log en MongoDB y responde al Frontend.
