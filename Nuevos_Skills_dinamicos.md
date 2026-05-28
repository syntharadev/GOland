# Habilidades Dinámicas del Sistema

- **Skill_001 [Protocolo de Comunicación UI-Backend]:**
  - **Trigger:** La UI necesita renderizar estados múltiples de agentes concurrentemente.
  - **Regla Adoptada:** Toda comunicación asíncrona hacia el frontend debe realizarse obligatoriamente a través del túnel establecido en `/ws/swarm`. Las rutas HTTP tradicionales (`GET/POST`) quedan relegadas únicamente a autenticación, carga de configuraciones iniciales (RAG/Narrativa) y comprobaciones de estado.

- **Skill_004 [Arquitectura de Mensajería Event-Driven]:**
  - **Trigger:** Conexión del frontend con el motor IA.
  - **Regla Adoptada:** Toda comunicación por el WebSocket debe estar estrictamente tipada con un campo `tipo` (ej. `INIT_WORLD`, `WORLD_READY`, `EVALUATE_CODE`). Ni el backend ni el frontend cuando enviar texto plano no estructurado. Todo evento viaja envuelto en JSON para que la UI sepa exactamente qué panel reactivo debe animar o actualizar.

- **Skill_005 [Renderizado Dinámico Basado en JSON]:**
  - **Trigger:** Evento entrante `WORLD_READY` vía WebSocket.
  - **Regla Adoptada:** El frontend tiene prohibido renderizar layouts estáticos para los agentes. El DOM de la tarjeta del personaje (`.gonion-card`) debe ser generado iterativamente leyendo las propiedades mutables inyectadas por Gemini (`nombre`, `personalidad`, `apariencia_sugerida`). Esto asegura total adaptabilidad a cualquier temática de usuario.

- **Skill_006 [Delegación de Voz del Consejo]:**
  - **Trigger:** Evaluación de código.
  - **Regla Adoptada:** El Orquestador ya no habla en nombre propio al evaluar. Debe inyectar obligatoriamente el nombre de un agente específico (previamente generado en `WORLD_READY`) en el campo `gonion_interviniente` para mantener la inmersión narrativa de que un experto particular está corrigiendo al usuario.

- **Skill_007 [Transiciones Cinemáticas Dirigidas por Estado]:**
  - **Trigger:** Cambio en la fase del usuario (Onboarding -> Generación -> Juego).
  - **Regla Adoptada:** El frontend debe forzar el foco (`flex` expansion) programáticamente hacia el panel que requiere la atención del usuario en ese milisegundo exacto, imitando el trabajo de un director de cámara. Cuando la IA "piensa", el foco va al Terminal. Cuando la IA "presenta", el foco va a los Agentes. Cuando el usuario debe "actuar", el foco va al Editor.

- **Skill_008 [Arranque Fluido "Zero-Click" de Nivel]:**
  - **Trigger:** Renderizado de `WORLD_READY`.
  - **Regla Adoptada:** El frontend asume la responsabilidad de la atención. Una vez inyectado el `RetoInicial`, el DOM ejecuta `document.querySelector('.panel-editor').focus()`. Esto desplaza las físicas CSS hacia el editor automáticamente, guiando la vista del usuario exactamente a donde debe escribir sin requerir un clic explicativo.

- **Skill_010 [Persistencia Event-Driven]:**
  - **Trigger:** Evento de aprobación de código.
  - **Regla Adoptada:** La base de datos es un consumidor pasivo. El flujo de juego no se detiene a esperar la confirmación de la base de datos; el guardado ocurre en la misma goroutine que transmite la victoria al usuario.

- **Skill_011 [Generación de Mundos Contextuales Mutables]:**
  - **Trigger:** Llamada a `GenerateSwarmWorld` con un `nivelActual > 1`.
  - **Regla Adoptada:** El Orquestador ahora tiene la capacidad de "actuar" como si ya conociera al usuario. El *System Prompt* requiere que los GOnions den la bienvenida al usuario reconociendo explícitamente su retorno en el `mensaje_gonion`, manteniendo intacta la cuarta pared narrativa tras una recarga de página.
