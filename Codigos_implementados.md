# Registro de Códigos Implementados

## `cmd/server/main.go`
- **Qué:** Punto de entrada de la aplicación y configuración del servidor HTTP.
- **Cómo:** Usa `http.NewServeMux()` para el enrutamiento (aprovechando las nuevas capacidades de método en Go 1.22 como `"GET /ruta"`). Define timeouts estrictos en la estructura `http.Server`.
- **Por qué:** Evita dependencias externas como Gin o Fiber para mantener el binario ligero. Los timeouts (`ReadTimeout`, `WriteTimeout`) previenen fugas de recursos (Slowloris attacks).

## `internal/api/handlers.go`
- **Qué:** Controladores de rutas, específicamente el Upgrade de HTTP a WebSocket.
- **Cómo:** Usa `gorilla/websocket` con un `Upgrader` que intercepta la petición `/ws/swarm` y la transforma en un canal bidireccional, manteniendo un bucle `for` de lectura y escritura.
- **Por qué:** El protocolo HTTP estándar es ineficiente para el streaming de datos en tiempo real (necesario para ver a los agentes "pensar"). WebSocket provee la persistencia necesaria con baja latencia. El `CheckOrigin = true` es una concesión temporal para desarrollo local.

## `cmd/server/main.go` (Actualización Inyección de Dependencias)
- **Qué:** Se inyectó el cliente de `geminiClient` dentro del enrutador de WebSockets.
- **Cómo:** Usando una función anónima (*closure*) en `mux.HandleFunc` para pasar el cliente inicializado al handler sin usar variables globales.
- **Por qué:** Mantiene el estado limpio y seguro, asegurando que la conexión a la API se inicializa una sola vez al arrancar y se comparte de forma segura entre todas las conexiones concurrentes.

## `internal/api/handlers.go` (Actualización Concurrencia)
- **Qué:** Integración de la lógica de negocio (Gemini) con la capa de red (WebSocket).
- **Cómo:** Se definió un `struct` `MensajeEntrante` para parsear la intención del usuario. Si el tipo es `INIT_WORLD`, se lanza una `goroutine` anónima (`go func()`) que maneja la petición RAG.
- **Por qué:** Si el usuario solicita un mundo y Gemini tarda 3 segundos en responder, bloquear el bucle `for` impediría que el servidor lea otros mensajes (como un botón de "cancelar" o un "ping" de conexión). La goroutine asegura una comunicación fluida asíncrona.

## `cmd/server/main.go` (Actualización Servidor de Estáticos)
- **Qué:** Inyección del gestor de archivos estáticos en el multiplexor de Go.
- **Cómo:** Usando `http.FileServer(http.Dir("./ui"))` acoplado a la ruta raíz `"/"`. El enrutador nativo de Go prioriza los matches exactos (`/ws/swarm`, `/health`) y deriva cualquier otra petición de archivo al servidor de estáticos.
- **Por qué:** Elimina la necesidad de correr servidores separados para frontend (como Node/Vite) durante el desarrollo local, unificando todo bajo el puerto único `:8080` y previniendo desajustes de red.

## `ui/index.html`, `style.css`, `app.js` (Estructura Agentic UI v1)
- **Qué:** Esqueleto de la interfaz reactiva basada en eventos del enjambre de agentes.
- **Cómo:** El CSS implementa `transition: flex 0.4s cubic-bezier(...)` sobre la clase `.panel`. Al disparar los estados `:hover` o `:focus-within`, la proporción cambia de `flex:1` a `flex:2.2`. JavaScript gestiona la serialización nativa del WebSocket e intercepta las transiciones de estado (`status`) y renderizado final (`WORLD_READY`).
- **Por qué:** Cumple con la directiva de diseño premium de competiciones: una interfaz interactiva fluida donde la IA alimenta los paneles asíncronamente sin recargar la página.

## `internal/llm/evaluator.go`
- **Qué:** Motor de evaluación de código impulsado por IA.
- **Cómo:** Crea una estructura `Evaluacion` con un booleano de aprobación y feedback narrativo. Configura el modelo para actuar como compilador/auditor estricto. Recibe el código raw del usuario y el objetivo del nivel, y devuelve una decisión estructurada.
- **Por qué:** Permite crear un bucle de gamificación seguro. Evitamos los enormes riesgos de seguridad de compilar y ejecutar código ajeno en nuestro backend (RCE), delegando el "entendimiento" del código al modelo multimodal de Gemini, que puede detectar errores de sintaxis y lógica con alta precisión.

## `internal/api/handlers.go` (Actualización de Eventos)
- **Qué:** Nuevo interceptor en el switch del WebSocket para el evento `EVALUATE_CODE`.
- **Cómo:** Lanza una nueva goroutine que invoca `gemini.EvaluateCode` y emite el resultado bajo el evento `CODE_EVALUATED`.
- **Por qué:** Mantiene la arquitectura orientada a eventos no bloqueante. La UI sabrá cuándo pintar una alerta verde (Aprobado) o roja (Rechazado) basada en este JSON tipado.

*Nota: Durante esta fase de auditoría UX (Fase 5.1), no se ha introducido código nuevo, sino que se ha establecido la especificación de la Máquina de Estados de la Interfaz (UI State Machine) basada en los eventos del WebSocket (`INIT_WORLD`, `WORLD_READY`, `EVALUATE_CODE`).*

## `ui/index.html` & `app.js` (Onboarding On-Rails)
- **Qué:** Implementación del Perfil de Usuario y Panel de Editor.
- **Cómo:** Se añadieron inputs para `Nick` y un selector dropdown para `NivelPrevio`. Se añadió un 4º panel reactivo para el Editor de Código que recibe las instrucciones del backend.
- **Por qué:** Recoger el nombre y el nivel permite personalizar el *prompt* de Gemini para ajustar la dificultad sintáctica del código y humanizar a los agentes llamando al usuario por su nombre, mitigando el GAP de diseño identificado.

## `internal/llm/orchestrator.go` (Inyección de Reto Cero)
- **Qué:** Modificación del RAG inicial para entregar la primera misión inmediatamente.
- **Cómo:** Se agregó el struct `RetoInicial` anidado dentro de `SwarmConfiguration`. El System Prompt se ajustó dinámicamente con `fmt.Sprintf` instruyendo explícitamente a Gemini para que genere la primera tarea basándose en la variable de experiencia (`Principiante`, `Intermedio`, `Experto`).
- **Por qué:** Consolida el ciclo de arranque. Ahora, una única llamada a la API inicializa el universo, crea los agentes y entrega el código base de la primera prueba, evitando tener que hacer un viaje de red extra para pedir el Nivel 1.

## `internal/db/repository.go`
- **Qué:** Módulo de persistencia local.
- **Cómo:** Implementa SQLite embebido usando `modernc.org/sqlite`. Ejecuta un `CREATE TABLE IF NOT EXISTS` al arrancar e incluye un método `SaveProgress` con lógica UPSERT (`ON CONFLICT... DO UPDATE`).
- **Por qué:** Evita requerir a los jueces instalar PostgreSQL localmente para probar el MVP. El UPSERT garantiza que un mismo usuario actualice su nivel sin generar duplicados.

## `cmd/server/main.go` & `handlers.go`
- **Qué:** Inyección de la base de datos en el ciclo de vida del servidor.
- **Cómo:** El handler ahora acepta `database *db.Repository`. El guardado se dispara asíncronamente solo cuando `eval.Aprobado` es `true`.
- **Por qué:** Separación de responsabilidades. El LLM evalúa, la base de datos persiste, el WebSocket transmite.

## Sistema de Recuperación de Sesiones (`handlers.go` & `repository.go`)
- **Qué:** Mecanismo de "Smart Login" transparente.
- **Cómo:** El handler HTTP intercepta `INIT_WORLD`, consulta a SQLite vía `GetUserProgress`. Si el Nick existe, sobreescribe las variables en memoria con las de la DB y se las pasa al Orquestador RAG (`GenerateSwarmWorld`) inyectando la variable `nivelActual`.
- **Por qué:** Permite retener a los usuarios y evitar sistemas complejos de autenticación (JWT/OAuth) en la fase MVP, manteniendo un UX fricción-cero.

## Testing de Seguridad (`repository_test.go`)
- **Qué:** Suite de pruebas funcionales y de inyección.
- **Cómo:** Instancia SQLite localmente usando `:memory:` para no ensuciar la base de datos `goland.db` del servidor real. Prueba los métodos CRUD y fuerza strings maliciosos (`' OR '1'='1`) para auditar la sanitización.
- **Por qué:** Demuestra ingeniería de software robusta, validando la estabilidad de los datos y protegiendo el backend antes del despliegue público.
