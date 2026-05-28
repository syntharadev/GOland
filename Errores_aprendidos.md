# Bitácora de Errores y Soluciones

*NOTA: En esta Fase 1, no se han registrado fallos de compilación ni ejecución durante la inicialización base. Se establece la estructura inicial para monitoreo.*

- **Prevención de Error [Seguridad - Timeouts HTTP]:** 
  - **Contexto:** Usar `http.ListenAndServe(":8080", nil)` expone el servidor a fugas de memoria si las conexiones se quedan colgadas.
  - **Solución implementada:** Se ha instanciado un `http.Server` personalizado en `main.go` declarando `ReadTimeout`, `WriteTimeout` e `IdleTimeout`.
  - **Justificación:** Las llamadas a APIs LLM pueden tener latencia. Proteger los descriptores de archivo del servidor backend es crítico para la escalabilidad.

- **Prevención de Error [Bloqueo de Hilo en WebSockets]:**
  - **Contexto:** Llamar directamente a `gemini.GenerateSwarmWorld` dentro del bucle `for` de lectura del WebSocket detiene por completo la escucha de nuevos eventos del cliente hasta que la API responde.
  - **Solución implementada:** Se encapsuló la llamada al LLM y la respuesta de red dentro de una *goroutine* (`go func() {...}()`).
  - **Justificación técnica:** En Go, los WebSockets deben tener canales de lectura/escritura concurrentes. Delegar el I/O pesado (llamadas de red externas) a una goroutine previene cuellos de botella en la conexión (Head-of-line blocking).

- **Prevención de Error [MIME Types Incorrectos en Servidores Go]:**
  - **Contexto:** Si los archivos estáticos se leen erróneamente usando rutas absolutas rotas, el navegador rechaza el CSS por seguridad MIME type (`text/plain`).
  - **Solución aplicada:** Se configuró el directorio relativo nativo `./ui` en Go y se enlazó el estilo en el HTML usando `<link rel="stylesheet" href="style.css">` de forma puramente relativa.
  - **Justificación técnica:** `http.FileServer` de Go deduce automáticamente los MIME content-types correctos (`text/css`, `application/javascript`) si lee los archivos nativos directamente desde la raíz del árbol jerárquico.

- **Prevención de Error [Ejecución Remota de Código (RCE)]:**
  - **Contexto:** Diseñar un backend que reciba un string, lo guarde en un archivo `.go` y ejecute `go run` expone el servidor a ataques masivos si un usuario malicioso escribe código que borra el disco o abre puertos (Reverse Shells).
  - **Solución aplicada:** Se implementó *Análisis Estático Generativo*. El código nunca se compila en nuestra máquina; solo se analiza como texto por la API de Gemini.
  - **Justificación técnica:** Sacrifica la ejecución real en favor de un 100% de seguridad en la fase MVP, manteniendo una experiencia de aprendizaje idéntica para el usuario, ya que la IA es capaz de predecir la salida del programa con exactitud.

- **Prevención de Error [UX Dead-End / Callejón sin salida]:**
  - **Contexto:** Tras la auditoría del flujo de usuario, descubrimos que después de recibir el evento `WORLD_READY`, la aplicación no instruye al usuario sobre qué debe programar. El panel de evaluación está inactivo porque no hay un "Objetivo de Nivel 1" definido.
  - **Solución Propuesta:** Modificar el Orquestador o crear un nuevo método `GenerateLevelTask(nivel int, contexto string)` que dispare el primer desafío justo después de presentar a los GOnions.
  - **Justificación técnica:** Una gamificación efectiva requiere un CTA (Call to Action) constante. El usuario nunca debe preguntarse "¿Y ahora qué hago?".

- **Prevención de Error [Desincronización de Dificultad]:**
  - **Contexto:** Si no se pasa el nivel de experiencia del usuario al Orquestador, un usuario principiante que elija "Ciberseguridad" podría recibir como primera tarea "Escribe un escáner de puertos concurrente", provocando frustración y abandono inmediato.
  - **Solución implementada:** Parámetro restrictivo en el Prompt. Si es Principiante, la IA fuerza un reto de variables/impresión tematizado (ej. "Imprime 'Sistema de Defensa Activo'").
  - **Justificación técnica:** Gamificación adaptativa. El modelo LLM debe actuar como un profesor empático, no como un compilador implacable.

- **Prevención de Error [CGO Build Failures]:**
  - **Contexto:** Usar `github.com/mattn/go-sqlite3` es estándar, pero requiere CGO y un compilador GCC instalado, lo que rompe despliegues rápidos en contenedores o máquinas limpias.
  - **Solución aplicada:** Se adoptó `modernc.org/sqlite`, un port puro en Go.
  - **Justificación técnica:** Asegura que `go build` funcione en cualquier arquitectura sin dependencias externas, crucial para entregar un binario limpio en competiciones.

- **Prevención de Error [Inconsistencia de Estado UI/Backend]:**
  - **Contexto:** Si el backend recupera el Nivel 5 de la base de datos, pero el frontend asume que es el Nivel 1 en su variable `appState`, la siguiente evaluación enviaría "Nivel 1" al evaluador, corrompiendo la progresión.
  - **Solución implementada:** El payload JSON de `WORLD_READY` se modificó para incluir el nodo `nivel_recuperado`. `app.js` lee esta variable y sincroniza su estado local (`appState.nivelActual = response.nivel_recuperado`) antes de renderizar el mundo.
  - **Justificación técnica:** El paradigma "Single Source of Truth". El backend dicta el estado real; el frontend lo refleja.

- **Fallo de Procedimiento [Omisión de Documentación de Secretos Locales]:**
  - **Contexto:** Se instruyó al usuario sobre la creación de un archivo `.env.example`, pero se omitió el paso explícito de clonarlo a `.env` y rellenarlo antes de la ejecución local, causando confusión sobre la gestión de claves y costes.
  - **Causa Raíz:** Asunción errónea del nivel de familiaridad del usuario con los estándares de inyección de entorno en Go.
  - **Solución adoptada:** Se ha establecido como regla documentar explícitamente la creación del `.env` local y se ha protocolizado el uso exclusivo del "Free Tier" de Google AI Studio para fases MVP, aislando al desarrollador de riesgos de facturación.
