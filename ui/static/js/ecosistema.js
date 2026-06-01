document.addEventListener('DOMContentLoaded', () => {
    const chatInput = document.getElementById('chat-input-text');
    const sendBtn = document.getElementById('btn-send-chat');
    const chatHistory = document.getElementById('chat-history');
    const roadmapContainer = document.getElementById('roadmap-container');
    const evaluateBtn = document.getElementById('btn-evaluar-code');
    const codeEditor = document.getElementById('workspace-editor');

    function scrollToBottom() {
        if (chatHistory) {
            chatHistory.scrollTop = chatHistory.scrollHeight;
        }
    }

    // Función genérica para inyectar mensajes con avatares y nombres
    function appendAgentMessage(name, text, avatarSrc, isUser = false) {
        if (!chatHistory) return;

        const bubbleWrapper = document.createElement('div');
        bubbleWrapper.className = 'chat-bubble-wrapper';

        // Burbuja principal
        const bubble = document.createElement('div');
        bubble.className = `chat-bubble ${isUser ? 'right' : 'left'}`;

        // Nombre del Agente arriba de la burbuja
        const nameLabel = document.createElement('div');
        nameLabel.className = 'chat-bubble-agent-name';
        nameLabel.innerText = name;

        bubble.innerHTML = `
            <img src="${avatarSrc}" alt="${name}" class="chat-avatar-thumb">
            <div class="chat-bubble-content">${text}</div>
        `;

        bubbleWrapper.appendChild(nameLabel);
        bubbleWrapper.appendChild(bubble);
        chatHistory.appendChild(bubbleWrapper);
        scrollToBottom();
    }

    // Acto 1: Carga de página y bienvenida del Profesor tras 1 segundo
    setTimeout(() => {
        appendAgentMessage(
            'El Profesor',
            '¡Bienvenido a GOland! Soy el Profesor. ¿Qué tipo de herramienta o proyecto te gustaría construir hoy con Go? 🐹',
            '/static/img/Profesor.png',
            false
        );
    }, 1000);

    // Manejador para enviar el input del usuario
    async function handleSend() {
        const query = chatInput.value.trim();
        if (!query) return;

        // Inyectar el mensaje del estudiante inmediatamente
        appendAgentMessage('Estudiante', query, '/static/img/Estudiante_1.png', true);
        chatInput.value = '';

        // Bloquear input
        chatInput.disabled = true;
        sendBtn.disabled = true;

        // Mostrar indicador de carga animado ("Enjambre Pensando...")
        const thinkingId = 'thinking-' + Date.now();
        const thinkingWrapper = document.createElement('div');
        thinkingWrapper.className = 'chat-bubble-wrapper';
        thinkingWrapper.id = thinkingId;

        const nameLabel = document.createElement('div');
        nameLabel.className = 'chat-bubble-agent-name';
        nameLabel.innerText = 'El Enjambre';

        const thinkingBubble = document.createElement('div');
        thinkingBubble.className = 'chat-bubble left';
        thinkingBubble.innerHTML = `
            <img src="/static/img/Profesor.png" alt="Enjambre" class="chat-avatar-thumb" style="animation: pulse 1s infinite alternate;">
            <div class="chat-bubble-content" style="font-style: italic; color: #64748b;">
                Coordinando con el enjambre de GOnions... 🧠
            </div>
        `;
        thinkingWrapper.appendChild(nameLabel);
        thinkingWrapper.appendChild(thinkingBubble);
        chatHistory.appendChild(thinkingWrapper);
        scrollToBottom();

        try {
            // Acto 3: Llamar al endpoint local POST /api/generar-ruta
            const response = await fetch('/api/generar-ruta', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ proyecto: query }),
            });

            const data = await response.json();

            // Eliminar indicador de pensando
            const thinkingElem = document.getElementById(thinkingId);
            if (thinkingElem) thinkingElem.remove();

            // Desbloquear input para futuras preguntas
            chatInput.disabled = false;
            sendBtn.disabled = false;

            // Acto 4: Respuesta del enjambre orquestada por setTimeout
            // A los 500ms, la Bibliotecaria
            setTimeout(() => {
                appendAgentMessage(
                    'La Bibliotecaria',
                    '¡Excelente elección! He indexado la documentación necesaria en la base de datos para ese proyecto. Consúltame cuando te atasques.',
                    '/static/img/Bibliotecaria.png',
                    false
                );
            }, 500);

            // A los 1500ms, el Mensajero con links
            setTimeout(() => {
                const linksHtml = `
                    ¡Paquete entregado! He interceptado estos 3 enlaces de repositorios de referencia que te servirán de inspiración:<br>
                    <a href="https://github.com/go-telegram-bot-api/telegram-bot-api" target="_blank" style="color: #00add8; text-decoration: underline; font-weight: 600;">1. Telegram Bot API en Go</a><br>
                    <a href="https://go.dev/doc/tutorial/web-service-gin" target="_blank" style="color: #00add8; text-decoration: underline; font-weight: 600;">2. Servicios Web con Gin-Gonic</a><br>
                    <a href="https://github.com/stretchr/testify" target="_blank" style="color: #00add8; text-decoration: underline; font-weight: 600;">3. Suite de Pruebas Testify</a>
                `;
                appendAgentMessage(
                    'El Mensajero',
                    linksHtml,
                    '/static/img/Mensajero.png',
                    false
                );

                // Renderizar el Roadmap en el panel superior izquierdo
                renderizarRoadmap(data);
            }, 1800);

        } catch (error) {
            console.error('Error generando la ruta del enjambre:', error);
            const thinkingElem = document.getElementById(thinkingId);
            if (thinkingElem) thinkingElem.remove();
            
            chatInput.disabled = false;
            sendBtn.disabled = false;

            appendAgentMessage(
                'El Profesor',
                'Disculpa la interferencia cuántica. Parece que el canal de comunicación con el enjambre temporalmente falló. ¿Podrías intentar de nuevo? 🐹',
                '/static/img/Profesor.png',
                false
            );
        }
    }

    // Función para renderizar el Roadmap en el panel izquierdo
    function renderizarRoadmap(data) {
        if (!roadmapContainer) return;

        // Limpiar el contenido anterior de bienvenida/espera
        roadmapContainer.innerHTML = '';

        const misiones = data.misiones || [];
        if (misiones.length === 0) {
            roadmapContainer.innerHTML = `
                <div style="text-align: center; padding: 20px; color: var(--color-text-dim);">
                    No se pudieron cargar las misiones de tu proyecto. Intenta de nuevo.
                </div>
            `;
            return;
        }

        // Crear contenedor envoltura para animaciones individuales
        const roadmapWrapper = document.createElement('div');
        roadmapWrapper.className = 'roadmap-wrapper';

        misiones.forEach((mision, index) => {
            const card = document.createElement('div');
            card.className = 'mision-card';
            card.style.animationDelay = `${index * 150}ms`; // Efecto cascada de aparición

            card.innerHTML = `
                <div class="mision-left">
                    <span class="mision-concepto">Misión ${index + 1}: ${mision.concepto}</span>
                    <span class="mision-titulo">${mision.titulo}</span>
                    <p class="mision-desc">${mision.descripcion}</p>
                </div>
                <div class="mision-right">⚡</div>
            `;

            // Click interactivo en cada tarjeta de misión
            card.addEventListener('click', () => {
                seleccionarMision(mision, index + 1);
            });

            roadmapWrapper.appendChild(card);
        });

        roadmapContainer.appendChild(roadmapWrapper);
    }

    // Cargar detalles de misión en el editor
    function seleccionarMision(mision, numeroMision) {
        // Al hacer click, preparamos el editor y mandamos un mensaje del Profesor al chat
        appendAgentMessage(
            'El Profesor',
            `Excelente decisión. Has seleccionado la <strong>Misión ${numeroMision}: ${mision.titulo}</strong>. Revisa la lección y pon a prueba tu código en el panel derecho. 🐹`,
            '/static/img/Profesor.png',
            false
        );

        // Pre-cargar código base correspondiente en el editor
        if (codeEditor) {
            let baseCode = `package main\n\nimport "fmt"\n\nfunc main() {\n    // Misión ${numeroMision}: ${mision.concepto}\n`;
            if (numeroMision === 1) {
                baseCode += `    // Declara una variable llamada token con valor "12345" usando :=\n    token := "12345"\n    fmt.Println("Conectando con token:", token)\n}`;
            } else if (numeroMision === 2) {
                baseCode += `    // Estructura para procesar JSON de entrada\n    type Mensaje struct {\n        Texto string \`json:"texto"\`\n    }\n    \n    m := Mensaje{Texto: "Hola de nuevo"}\n    fmt.Println("Procesando:", m.Texto)\n}`;
            } else {
                baseCode += `    // Ejecutando en paralelo con Goroutines\n    ch := make(chan bool)\n    go func() {\n        fmt.Println("Acción cuántica paralela!")\n        ch <- true\n    }()\n    <-ch\n}`;
            }
            codeEditor.value = baseCode;
        }
    }

    // Listeners
    if (sendBtn) sendBtn.addEventListener('click', handleSend);
    if (chatInput) {
        chatInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') handleSend();
        });
    }

    if (evaluateBtn) {
        evaluateBtn.addEventListener('click', () => {
            alert('¡Enviando código del estudiante a La Guardiana para auditoría cuántica! 🛡️\nCompilando y ejecutando suite de pruebas...');
        });
    }
});
