document.addEventListener('DOMContentLoaded', () => {
    const chatInput = document.getElementById('chat-input-text');
    const sendBtn = document.getElementById('btn-send-chat');
    const chatHistory = document.getElementById('chat-history');
    const roadmapContainer = document.getElementById('roadmap-container');
    const evaluateBtn = document.getElementById('btn-evaluar-code');
    const codeEditor = document.getElementById('workspace-editor');

    let hasRoadmap = false;

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

    // Función para consultar a La Bibliotecaria RAG MCP MongoDB
    async function consultarBibliotecaria(pregunta) {
        // Bloquear input
        chatInput.disabled = true;
        sendBtn.disabled = true;

        // Mostrar indicador de carga animado
        const thinkingId = 'thinking-bibliotecaria-' + Date.now();
        const thinkingWrapper = document.createElement('div');
        thinkingWrapper.className = 'chat-bubble-wrapper';
        thinkingWrapper.id = thinkingId;

        const nameLabel = document.createElement('div');
        nameLabel.className = 'chat-bubble-agent-name';
        nameLabel.innerText = 'La Bibliotecaria';

        const thinkingBubble = document.createElement('div');
        thinkingBubble.className = 'chat-bubble left';
        thinkingBubble.innerHTML = `
            <img src="/static/img/Bibliotecaria.png" alt="Bibliotecaria" class="chat-avatar-thumb" style="animation: pulse 1s infinite alternate;">
            <div class="chat-bubble-content" style="font-style: italic; color: #64748b;">
                Buscando en la base de datos de MongoDB Atlas... 📚
            </div>
        `;
        thinkingWrapper.appendChild(nameLabel);
        thinkingWrapper.appendChild(thinkingBubble);
        chatHistory.appendChild(thinkingWrapper);
        scrollToBottom();

        try {
            const response = await fetch('/api/agentes/bibliotecaria', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ pregunta: pregunta }),
            });

            const data = await response.json();

            // Eliminar indicador
            const thinkingElem = document.getElementById(thinkingId);
            if (thinkingElem) thinkingElem.remove();

            // Desbloquear input
            chatInput.disabled = false;
            sendBtn.disabled = false;
            chatInput.focus();

            // Formatear Markdown básico
            let formattedText = data.respuesta
                .replace(/\n/g, '<br>')
                .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                .replace(/\*(.*?)\*/g, '<em>$1</em>')
                .replace(/`(.*?)`/g, '<code style="background: rgba(14, 165, 233, 0.08); padding: 2px 6px; border-radius: 4px; font-family: monospace;">$1</code>');

            appendAgentMessage('La Bibliotecaria', formattedText, '/static/img/Bibliotecaria.png', false);

        } catch (error) {
            console.error('Error al consultar a La Bibliotecaria:', error);
            const thinkingElem = document.getElementById(thinkingId);
            if (thinkingElem) thinkingElem.remove();

            chatInput.disabled = false;
            sendBtn.disabled = false;

            appendAgentMessage(
                'La Bibliotecaria',
                'Error cuántico: No se pudo conectar con la base de datos de MongoDB.',
                '/static/img/Bibliotecaria.png',
                false
            );
        }
    }

    // Manejador para enviar el input del usuario
    async function handleSend() {
        const query = chatInput.value.trim();
        if (!query) return;

        // Inyectar el mensaje del estudiante inmediatamente
        appendAgentMessage('Estudiante', query, '/static/img/Estudiante_1.png', true);
        chatInput.value = '';

        // Si ya hay un roadmap generado, canalizar la duda directamente a La Bibliotecaria
        if (hasRoadmap) {
            await consultarBibliotecaria(query);
            return;
        }

        // De lo contrario, iniciar el flujo de onboarding y generación de ruta
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
            hasRoadmap = true; // El roadmap ya está disponible

            // Acto 4: Respuesta del enjambre orquestada por setTimeout
            // A los 500ms, la Bibliotecaria
            setTimeout(() => {
                appendAgentMessage(
                    'La Bibliotecaria',
                    '¡Excelente elección! He indexado la documentación necesaria en la base de datos para ese proyecto. Consúltalos aquí o pregúntame directamente cualquier duda que tengas sobre Go. 📚',
                    '/static/img/Bibliotecaria.png',
                    false
                );
            }, 500);

            // A los 1800ms, el Mensajero con links
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
        hasRoadmap = true; // Activar el modo RAG al interactuar con el temario
        
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

    // Función para ejecutar código mediante El Constructor (GitLab CI/CD RAG)
    async function ejecutarCodigoEstudiante() {
        if (!codeEditor) return;
        const codigo_editor = codeEditor.value.trim();

        if (evaluateBtn) evaluateBtn.disabled = true;

        // Inyectar mensaje automático del Mensajero
        appendAgentMessage(
            'El Mensajero',
            'Enviando código al pipeline CI/CD de GitLab... 📦',
            '/static/img/Mensajero.png',
            false
        );

        // Mostrar indicador de carga animado en el chat flotante
        const loadingId = 'loading-constructor-' + Date.now();
        const loadingWrapper = document.createElement('div');
        loadingWrapper.className = 'chat-bubble-wrapper';
        loadingWrapper.id = loadingId;

        const nameLabel = document.createElement('div');
        nameLabel.className = 'chat-bubble-agent-name';
        nameLabel.innerText = 'El Constructor';

        const loadingBubble = document.createElement('div');
        loadingBubble.className = 'chat-bubble left';
        loadingBubble.innerHTML = `
            <img src="/static/img/Constructor.png" alt="Constructor" class="chat-avatar-thumb" style="animation: pulse 1s infinite alternate;">
            <div class="chat-bubble-content" style="font-style: italic; color: #64748b;">
                Ejecutando pipeline y compilando en Go... ⚙️
            </div>
        `;
        loadingWrapper.appendChild(nameLabel);
        loadingWrapper.appendChild(loadingBubble);
        chatHistory.appendChild(loadingWrapper);
        scrollToBottom();

        try {
            const response = await fetch('/api/agentes/constructor', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ codigo: codigo_editor }),
            });

            const data = await response.json();

            // Eliminar indicador
            const loadingElem = document.getElementById(loadingId);
            if (loadingElem) loadingElem.remove();
            if (evaluateBtn) evaluateBtn.disabled = false;

            // Determinar si la respuesta tiene errores para aplicar el color pastel
            const esError = data.respuesta.includes('failed') || data.respuesta.includes('Error') || data.respuesta.includes('❌');

            const bubbleWrapper = document.createElement('div');
            bubbleWrapper.className = 'chat-bubble-wrapper';

            const nameLabelFinal = document.createElement('div');
            nameLabelFinal.className = 'chat-bubble-agent-name';
            nameLabelFinal.innerText = 'El Constructor';

            const bubble = document.createElement('div');
            bubble.className = 'chat-bubble left';

            // Resaltado de estilos según el status de compilación
            const styleColor = esError 
                ? 'background: rgba(254, 242, 242, 0.95); border: 1.5px solid rgba(239, 68, 68, 0.2); color: #991b1b; font-family: monospace;' 
                : 'background: rgba(240, 253, 250, 0.95); border: 1.5px solid rgba(16, 185, 129, 0.2); color: #065f46; font-family: monospace;';

            let formattedText = data.respuesta.replace(/\n/g, '<br>');

            bubble.innerHTML = `
                <img src="/static/img/Constructor.png" alt="Constructor" class="chat-avatar-thumb">
                <div class="chat-bubble-content" style="${styleColor}">${formattedText}</div>
            `;

            bubbleWrapper.appendChild(nameLabelFinal);
            bubbleWrapper.appendChild(bubble);
            chatHistory.appendChild(bubbleWrapper);
            scrollToBottom();

            // Si el pipeline es exitoso, invocar automáticamente La Cronometradora para telemetría APM
            if (!esError) {
                setTimeout(async () => {
                    const chronoLoadingId = 'loading-cronometradora-' + Date.now();
                    const chronoWrapper = document.createElement('div');
                    chronoWrapper.className = 'chat-bubble-wrapper';
                    chronoWrapper.id = chronoLoadingId;

                    const chronoName = document.createElement('div');
                    chronoName.className = 'chat-bubble-agent-name';
                    chronoName.innerText = 'La Cronometradora';

                    const chronoBubble = document.createElement('div');
                    chronoBubble.className = 'chat-bubble left';
                    chronoBubble.innerHTML = `
                        <img src="/static/img/Cronometradora.png" alt="La Cronometradora" class="chat-avatar-thumb" style="animation: pulse 1s infinite alternate;">
                        <div class="chat-bubble-content" style="font-style: italic; color: #64748b;">
                            Indexando telemetría en Elastic APM y analizando Big O... ⏱️
                        </div>
                    `;
                    chronoWrapper.appendChild(chronoName);
                    chronoWrapper.appendChild(chronoBubble);
                    chatHistory.appendChild(chronoWrapper);
                    scrollToBottom();

                    try {
                        const chronoRes = await fetch('/api/agentes/cronometradora', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({ codigo: codigo_editor }),
                        });

                        const chronoData = await chronoRes.json();

                        const chronoLoadingElem = document.getElementById(chronoLoadingId);
                        if (chronoLoadingElem) chronoLoadingElem.remove();

                        const chronoBubbleWrapper = document.createElement('div');
                        chronoBubbleWrapper.className = 'chat-bubble-wrapper';

                        const chronoNameLabel = document.createElement('div');
                        chronoNameLabel.className = 'chat-bubble-agent-name';
                        chronoNameLabel.innerText = 'La Cronometradora';

                        const finalBubble = document.createElement('div');
                        finalBubble.className = 'chat-bubble left';

                        const styleBlue = 'background: rgba(240, 249, 255, 0.95); border: 1.5px solid rgba(56, 189, 248, 0.2); color: #0369a1;';

                        let formattedChronoText = chronoData.respuesta
                            .replace(/\n/g, '<br>')
                            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                            .replace(/\*(.*?)\*/g, '<em>$1</em>');

                        finalBubble.innerHTML = `
                            <img src="/static/img/Cronometradora.png" alt="La Cronometradora" class="chat-avatar-thumb">
                            <div class="chat-bubble-content" style="${styleBlue}">${formattedChronoText}</div>
                        `;

                        chronoBubbleWrapper.appendChild(chronoNameLabel);
                        chronoBubbleWrapper.appendChild(finalBubble);
                        chatHistory.appendChild(chronoBubbleWrapper);
                        scrollToBottom();

                    } catch (err) {
                        console.error('Error con La Cronometradora:', err);
                        const chronoLoadingElem = document.getElementById(chronoLoadingId);
                        if (chronoLoadingElem) chronoLoadingElem.remove();

                        appendAgentMessage(
                            'La Cronometradora',
                            'No se pudo recuperar la telemetría APM de Elastic Search.',
                            '/static/img/Cronometradora.png',
                            false
                        );
                    }
                }, 1000);
            }

        } catch (error) {
            console.error('Error al ejecutar código con El Constructor:', error);
            const loadingElem = document.getElementById(loadingId);
            if (loadingElem) loadingElem.remove();
            if (evaluateBtn) evaluateBtn.disabled = false;

            appendAgentMessage(
                'El Constructor',
                'Error crítico: No se pudo conectar con el pipeline de GitLab CI/CD.',
                '/static/img/Constructor.png',
                false
            );
        }
    }

    // Listeners
    if (sendBtn) sendBtn.addEventListener('click', handleSend);
    if (chatInput) {
        chatInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') handleSend();
        });
    }

    // Evaluar botón
    if (evaluateBtn) {
        evaluateBtn.addEventListener('click', ejecutarCodigoEstudiante);
    }

    // Ataque de teclado Ctrl + Enter en el Editor de Código
    if (codeEditor) {
        codeEditor.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'Enter') {
                e.preventDefault();
                ejecutarCodigoEstudiante();
            }
        });
    }
});
