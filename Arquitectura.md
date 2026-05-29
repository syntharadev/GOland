# Arquitectura del Sistema y Despliegue de Producción (Vivo)

## 📌 1. Introducción y Estado de Producción
GOland ha completado con éxito la transición de un entorno MVP local a un entorno de **producción en la nube de alta disponibilidad, 100% serverless, inmutable y persistente**.

El sistema ha sido diseñado bajo los principios de **Zero Trust, escalabilidad de coste cero en inactividad y rendimiento optimizado**. Esta arquitectura permite que la experiencia de aprendizaje gamificada de GOland funcione a gran escala sin mantenimiento de servidores físicos y de forma totalmente automatizada.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                ESTADO                                   │
│  🟢 Producción Live & Operativa | Versión: 1.0.0 (MVP Completado)       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 🛠️ 2. Ficha Técnica de la Infraestructura Inmutable

A continuación se detalla la configuración inmutable que da soporte a GOland en su entorno productivo:

| Componente | Tecnología | Configuración / Detalles | Razón Estratégica |
| :--- | :--- | :--- | :--- |
| **Backend Core** | Go 1.26.3 | Multiplexor nativo (`http.NewServeMux`) y WebSockets con `gorilla/websocket`. | Rendimiento y concurrencia máxima con baja latencia y consumo de recursos mínimo sin frameworks pesados. |
| **Persistencia** | Supabase (PostgreSQL) | Conexión pooling mediante `pgx` (PostgreSQL Driver) usando la variable `DATABASE_URL`. | Base de datos relacional robusta en la nube. Supabase ofrece escalabilidad y compatibilidad nativa con SQL en su capa gratuita. |
| **Autenticación** | Google OAuth 2.0 | Flujo completo gestionado con redirecciones a callbacks seguras y cookies encriptadas (`gorilla/sessions`). | Autenticación robusta y segura de usuario sin fricción de contraseñas, validando tokens e impidiendo ataques de spoofing y CSRF. |
| **Alojamiento Cloud** | Google Cloud Run (Gen 2) | Serverless, escalado automático de **0 a 5 instancias** y **CPU Boost activado**. | Maximiza la capa gratuita de GCP (reducción a 0 instancias si no hay tráfico). CPU Boost elimina el impacto del cold-start al arrancar. |
| **Pipeline CI/CD** | Google Cloud Build | Integrado directamente con la rama `main` de GitHub. | Integración y despliegue continuos (GitOps). Cualquier push a `main` desencadena la validación, construcción y despliegue automático. |
| **Entorno Docker** | Dockerfile Multi-etapa | **Builder:** Go 1.26.3 Alpine<br>**Runner:** Alpine latest con `ca-certificates`. | Genera una imagen inmutable ultra-ligera (menos de 30MB) y segura. `ca-certificates` es mandatorio para llamadas SSL salientes a Gemini. |

---

## 📐 3. Diagrama de Arquitectura de Flujo de Datos

El siguiente diagrama ilustra cómo interactúan los componentes en el entorno de producción vivo. Muestra la comunicación en tiempo real entre el cliente (Agentic UI), el balanceador de Cloud Run, el backend en Go, y los servicios externos de terceros (Gemini, Google OAuth y Supabase).

```mermaid
graph TD
    %% Estilos de nodos
    classDef client fill:#EBF5FB,stroke:#2980B9,stroke-width:2px;
    classDef cloud fill:#EAFAF1,stroke:#27AE60,stroke-width:2px;
    classDef external fill:#FEF9E7,stroke:#D4AC0D,stroke-width:2px;
    classDef database fill:#FDEDEC,stroke:#C0392B,stroke-width:2px;

    %% Nodos
    User["🌐 Cliente (Browser/UI)"]:::client
    GCR["☁️ Google Cloud Run (Gen 2)<br>Backend Go 1.26.3"]:::cloud
    OAuth["🔑 Google OAuth 2.0"]:::external
    Gemini["🧠 Google Gemini API<br>(AI Studio)"]:::external
    Supabase["🗄️ Supabase PostgreSQL<br>(pgx Driver)"]:::database

    %% Relaciones
    User -->|1. Petición HTTPS / Assets Estáticos| GCR
    User -->|2. Flujo de Login Google| OAuth
    OAuth -->|3. Callback con Perfil| GCR
    User <-->|4. Canal WebSocket /ws/swarm| GCR
    GCR -->|5. Consultas SQL / Progreso (UPSERT)| Supabase
    GCR -->|6. RAG Prompt & Evaluación Estática| Gemini

    %% Leyendas
    subgraph Google Cloud Platform (GCP)
        GCR
    end
```

---

## 🔒 4. Variables de Entorno en Producción
Para garantizar la inmutabilidad y la seguridad, todas las credenciales sensibles están **aisladas del código fuente** y se inyectan en tiempo de ejecución en el contenedor de Google Cloud Run:

> [!IMPORTANT]
> Bajo ningún concepto se deben incluir credenciales en claro en los archivos del repositorio. Utilice la inyección del entorno de GCP para actualizar estas claves.

*   `GEMINI_API_KEY`: Clave de acceso a la API del modelo fundacional (Google AI Studio) para el Orquestador RAG y el Evaluador de código.
*   `SESSION_SECRET_KEY`: Semilla criptográfica utilizada por `gorilla/securecookie` y `gorilla/sessions` para encriptar y firmar las cookies de sesión en el navegador del cliente.
*   `GOOGLE_CLIENT_ID`: Identificador de cliente único asignado en Google Cloud Console para la API de Google OAuth 2.0.
*   `GOOGLE_CLIENT_SECRET`: Clave privada secreta de la API de Google OAuth 2.0 para el backend.
*   `DATABASE_URL`: Cadena de conexión JDBC/URI de PostgreSQL para Supabase (esquema `postgresql://user:pass@host:port/dbname`). Leída por la biblioteca de conexión nativa `pgx` para la persistencia transaccional y consistente del progreso del usuario.

---

## 🚀 5. Pipeline de CI/CD e Integración Continua (GitOps)

El despliegue se automatiza mediante un flujo GitOps acoplado a **Google Cloud Build**.

```mermaid
sequenceDiagram
    autonumber
    actor Dev as Desarrollador
    participant GH as GitHub (Branch main)
    participant CB as Google Cloud Build
    participant AR as Artifact Registry
    participant CR as Cloud Run (Gen 2)

    Dev->>GH: Git Push a la rama principal (main)
    GH->>CB: Webhook dispara Build Trigger
    Note over CB: Ejecuta Dockerfile Multi-Etapa
    CB->>CB: Etapa 1: Builder (Go 1.26.3 Alpine) - Compilación estática
    CB->>CB: Etapa 2: Runner (Alpine + ca-certificates) - Imagen final ultra-ligera
    CB->>AR: Sube la imagen Docker etiquetada (tag: latest)
    CB->>CR: Despliega nueva versión inmutable
    Note over CR: Activa CPU Boost para el arranque inicial
    CR-->>Dev: MVP Actualizado y Vivo en Producción
```

### 🐳 Dockerfile Multi-Etapa Explicado
La inmutabilidad del despliegue se garantiza mediante el `Dockerfile` de dos etapas:
1.  **Etapa de Compilación (`builder`):** Utiliza la imagen oficial de Go `golang:1.26.3-alpine`. Ejecuta la compilación con las banderas `CGO_ENABLED=0 GOOS=linux` para generar un binario estático y auto-contenido, eliminando la necesidad de enlazadores dinámicos.
2.  **Etapa de Ejecución (`runner`):** Utiliza una imagen base limpia de `alpine:latest`. Instala el paquete esencial `ca-certificates` (indispensable para verificar los certificados HTTPS de las llamadas API salientes a Gemini). Copia el binario estático `/app/goland-server` y la carpeta del frontend `/app/ui`. El contenedor resultante tiene un peso insignificante, lo que permite un despliegue ultra veloz de las réplicas en Cloud Run.

---

## 📈 6. Configuración de Alta Disponibilidad y Free Tier

Para optimizar recursos sin comprometer la velocidad de respuesta, el servicio en Google Cloud Run (Gen 2) se ha parametrizado bajo las siguientes directivas operacionales:

*   **Autoscaling 0 a 5 instancias:** La escala mínima en `0` asegura que no se consuma CPU ni memoria cuando no haya tráfico activo, lo que reduce la facturación a **cero absoluto** durante la inactividad. El tope máximo de `5` instancias concurrentes evita desbordamientos de facturación o ataques de denegación de servicio (DoS) imprevistos, conteniendo el MVP en la capa gratuita.
*   **CPU Boost Activado:** Al activarse una nueva instancia desde cero (debido a una petición entrante que saca al servicio del estado "frío"), Cloud Run asigna temporalmente más potencia de CPU. Esto reduce a menos de una fracción de segundo la latencia de arranque (Cold Start), asegurando que el canal WebSocket se establezca inmediatamente.
*   **Gestión de Conexiones WebSocket:** Gracias a que Go implementa eficientemente la concurrencia nativa a través de Goroutines y multiplexores no bloqueantes, un solo contenedor ligero puede dar soporte a cientos de sockets Web persistentes sin degradación del rendimiento.

---

## 🔒 7. Seguridad por Diseño (Security by Design)
1.  **Protección de Datos:** Las credenciales y claves privadas nunca se almacenan en la base de datos ni en archivos del código fuente. Las variables de entorno son inyectadas encriptadas por el panel de control de GCP.
2.  **Sandboxing Generativo:** GOland no evalúa código ejecutándolo localmente en el servidor (lo que abriría la puerta a inyecciones RCE / ejecución remota de código). En su lugar, el sistema delega la evaluación estática y sintáctica al modelo multimodal de Gemini, actuando como un evaluador de RAG seguro.
3.  **Seguridad de Sockets y CSRF:** El canal WebSocket `/ws/swarm` exige validación y los endpoints HTTP autentican al usuario basándose en cookies encriptadas en el cliente (`gorilla/securecookie`), protegiendo al backend de llamadas maliciosas cruzadas.
