# Módulo 5: Comunidad - El Gremio

## 📌 1. Introducción y Concepto
**El Gremio** es el espacio social y colaborativo en tiempo real de GOland. El aprendizaje de Go no es un viaje solitario; los tripulantes pueden compartir sus códigos, reportar fallas en los laboratorios, solicitar consejos sobre concurrencia avanzada y responder a las consultas de otros programadores.

Este módulo se apoya en una arquitectura híbrida impulsada por los eventos en tiempo real de **Supabase Realtime** y notificaciones asíncronas de bajo coste transmitidas por **WebSockets** bajo la supervisión del agente de IA **Senior** (moderador) y el agente **Mensajero** (notificaciones).

```
┌────────────────────────────────────────────────────────────────────────┐
│                        FLUJO DE LA COMUNIDAD                           │
│  Nuevo Post ➔ Supabase Trigger ➔ WebSocket broadcast ➔ Alerta Kawaii   │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 🏛️ 2. Foros Colaborativos y Tablón de Anuncios
La interfaz de la comunidad se presenta como una cartelera holográfica interactiva de vidrio (Glassmorphism), donde se organizan dos flujos principales:

1.  **El Tablón del Gremio:** El canal de anuncios oficiales emitidos por el sistema o por los administradores sobre retos globales semanales (ej. *"El Gran Reto de Canales Concurrentes"*).
2.  **Los Foros de Discusión:** Hilos interactivos de preguntas y respuestas categorizadas por etiquetas técnicas (sintaxis, base de datos, testing, concurrencia).

---

## ⚡ 3. Infraestructura de Mensajería Bidireccional Asíncrona

El motor de tiempo real que mantiene comunicada a la tripulación opera mediante dos canales síncronos:

```mermaid
graph TD
    %% Estilos de nodos
    classDef client fill:#EBF5FB,stroke:#2980B9,stroke-width:2px;
    classDef cloud fill:#EAFAF1,stroke:#27AE60,stroke-width:2px;
    classDef supabase fill:#FDEDEC,stroke:#C0392B,stroke-width:2px;

    %% Nodos
    Client1["🌐 Tripulante A (UI)"]:::client
    Client2["🌐 Tripulante B (UI)"]:::client
    GCR["☁️ Backend Go (Cloud Run)"]:::cloud
    WSChannel["🔌 Canal WebSocket /ws/swarm"]:::cloud
    SupaRealtime["🗄️ Supabase Realtime DB"]:::supabase

    %% Conexiones
    Client1 -->|1. Inserta nuevo Post/Respuesta| SupaRealtime
    SupaRealtime -->|2. Trigger en base de datos| GCR
    GCR -->|3. Procesa evento y asocia al Mensajero| WSChannel
    WSChannel -->|4. Broadcast de Notificación en Vivo| Client2
    Note over Client2: El Mensajero hace una animación Kawaii<br/>y muestra la alerta flotante en el UI.
```

### El Rol del Mensajero (WebSocket Alert Engine)
El **Mensajero** es la voz oficial del WebSocket de la comunidad.
*   Cuando se realiza una publicación en el foro o se completa un reto global, el backend capta el evento de base de datos y lo retransmite por el túnel `/ws/swarm`.
*   El Mensajero intercepta la trama JSON en el cliente y dibuja una burbuja de notificación en vivo con una animación elástica Kawaii (Hover Kawaii) que se desvanece a los 4 segundos:
    ```
    [MENSAJERO]: ⚡ ¡El usuario 'GolangNinja' acaba de resolver la Quest de Concurrencia! 🚀 (+100 XP)
    ```

---

## 👵 4. El Rol del Agente Senior (El Moderador Inteligente)
El **Senior** es el agente veterano basado en la API de Gemini 1.5, encargado de velar por el orden técnico y didáctico de la comunidad:

*   **Moderación Activa:** Lee automáticamente los nuevos posts publicados en el foro. Si un post contiene código inapropiado, spam o la respuesta explícita exacta de una misión de la Academia (arruinando la gamificación), el Senior edita automáticamente la publicación ocultando el código bajo un bloque translúcido difuminado y dejando una advertencia empática.
*   **Mentoría Automatizada:** Si una pregunta técnica en el foro no recibe respuesta de usuarios reales tras 10 minutos, el Senior se activa de forma automática. Consulta el RAG de la academia (`internal/llm/client.go`), extrae la especificación oficial del problema y publica una respuesta de mentoría detallada y amable para guiar al tripulante bloqueado.

---

## 🏆 5. Sistema de Desbloqueo por Logros de la Comunidad
La participación activa en el Gremio desbloquea contenido e insignias RPG exclusivas para el perfil del usuario:

### Cómo Desbloquear al Agente "Senior" en tu Roster
1.  **Interacción de Comunidad:** Los usuarios que necesiten consejos en foros pueden marcar una respuesta aportada por otro tripulante como **"Solución Útil"**.
2.  **Guardado en Supabase:** Esta acción incrementa el contador de reputación (`reputacion_foro`) en la tabla de base de datos del usuario que respondió.
3.  **El Trigger del Desbloqueo:** Al alcanzar **5 "Soluciones Útiles"** validadas:
    *   Supabase dispara una actualización que cambia el estado del agente `Senior` a `desbloqueado = TRUE` en la tabla `agentes_usuario`.
    *   El **Mensajero** emite un mensaje global en vivo vía WebSocket felicitando al tripulante por su ascenso técnico a la categoría de Mentor de la Nave.
    *   El avatar del **Senior** queda permanentemente disponible en la consola flotante para dar soporte en cualquier módulo.

---

## 📊 6. Estructura de Tablas del Gremio en Supabase
```sql
-- Tabla para los hilos de los foros de la comunidad
CREATE TABLE IF NOT EXISTS foro_hilos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo VARCHAR(255) NOT NULL,
    contenido TEXT NOT NULL,
    autor_nick TEXT REFERENCES progreso_usuarios(nick) ON DELETE CASCADE,
    categoria VARCHAR(50) DEFAULT 'General',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

-- Respuestas dentro de los hilos
CREATE TABLE IF NOT EXISTS foro_respuestas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hilo_id UUID REFERENCES foro_hilos(id) ON DELETE CASCADE,
    contenido TEXT NOT NULL,
    autor_nick TEXT REFERENCES progreso_usuarios(nick) ON DELETE CASCADE,
    es_solucion_util BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

-- Reputación acumulada del usuario
ALTER TABLE progreso_usuarios ADD COLUMN IF NOT EXISTS reputacion_foro INTEGER DEFAULT 0;
```
