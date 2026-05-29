// Detecta dinámicamente la dirección del host para abrir el WebSocket
const ws = new WebSocket(`ws://${window.location.host}/ws/swarm`);

const btnInit = document.getElementById('btn-init');
const selectNivel = document.getElementById('nivel-previo');
const inputDominio = document.getElementById('dominio');
const inputObjetivo = document.getElementById('objetivo');
const terminalLogs = document.getElementById('terminal-logs');
const narrativaOutput = document.getElementById('narrativa-output');
const agentesContainer = document.getElementById('agentes-container');

function logToTerminal(text, type = 'system') {
    const line = document.createElement('div');
    line.className = `log-line ${type}`;
    line.innerText = `[${new Date().toLocaleTimeString()}] ${text}`;
    terminalLogs.appendChild(line);
    // Auto-scroll hacia abajo
    terminalLogs.scrollTop = terminalLogs.scrollHeight;
}

ws.onopen = () => {
    logToTerminal('Conexión bidireccional establecida con el núcleo GOland.', 'success');
};

ws.onclose = () => {
    logToTerminal('Conexión perdida con el servidor de Go.', 'system');
};

ws.onmessage = (event) => {
    const response = JSON.parse(event.data);
    
    // Intercepta actualizaciones de progreso del Enjambre
    if (response.status) {
        logToTerminal(response.status, 'status');
    }
    
    if (response.error) {
        logToTerminal(`ERROR CRÍTICO: ${response.error}`, 'system');
    }

    // Intercepta la estructura JSON final enviada desde Gemini
    if (response.tipo === 'WORLD_READY') {
        logToTerminal('¡Consenso alcanzado! Entorno de simulación listo.', 'success');
        
        // SINCRONIZACIÓN DE ESTADO
        if (response.nivel_recuperado) {
            appState.nivelActual = response.nivel_recuperado;
        }
        if (appState.nivelActual > 1) {
            logToTerminal(`Bienvenido de nuevo, ${appState.nick}. Restaurando sistema al Nivel ${appState.nivelActual}.`, 'status');
        }

        renderWorld(response.data);
    }
    
    if (response.tipo === 'CODE_EVALUATED') {
        const evalData = response.data;
        if (evalData.aprobado) {
            logToTerminal(`[${evalData.gonion_interviniente}] ¡Aprobado! ${evalData.feedback}`, 'success');
            appState.nivelActual++;
        } else {
            logToTerminal(`[${evalData.gonion_interviniente}] Error: ${evalData.feedback}`, 'system');
        }
    }
};

const appState = {
    nick: '',
    nivelActual: 1
};

// Check Auth Status on load
document.addEventListener('DOMContentLoaded', () => {
    fetch('/auth/status')
        .then(res => res.json())
        .then(data => {
            if (data.authenticated) {
                appState.nick = data.nick;
                document.getElementById('btn-google-login').style.display = 'none';
                const authStatus = document.getElementById('auth-status');
                authStatus.innerText = `Autenticado como: ${data.nick}`;
                authStatus.style.display = 'block';
                document.getElementById('setup-form').style.display = 'block';
            }
        })
        .catch(err => console.error('Auth check failed:', err));
});

btnInit.addEventListener('click', () => {
    const nick = appState.nick;
    const nivel_previo = selectNivel.value;
    const dominio = inputDominio.value.trim();
    const objetivo = inputObjetivo.value.trim();
    
    if (!nick) {
        alert('Debes iniciar sesión con Google primero.');
        return;
    }
    if (!dominio || !objetivo) {
        alert('Por favor, completa el ámbito y el objetivo.');
        return;
    }

    const payload = {
        tipo: 'INIT_WORLD',
        nick: nick,
        nivel_previo: nivel_previo,
        dominio: dominio,
        objetivo: objetivo
    };

    appState.nick = nick;

    logToTerminal(`Enviando directrices al Orquestador... Piloto: [${nick}]`, 'system');
    ws.send(JSON.stringify(payload));
});

function renderWorld(data) {
    // 1. Inyectar Título e Hilo Conductor generado por la IA
    narrativaOutput.innerHTML = `
        <h3 style="color:#38bdf8; margin-bottom:10px;">🎮 Misión: ${data.tema || data.dominio || 'Simulación'}</h3>
        <p style="font-size:0.9rem; line-height:1.5; color:#cbd5e1;">${data.hilo_conductor || data.objetivo || 'Generando hilo narrativo...'}</p>
        <div style="margin-top:15px; font-size:0.8rem; color:#94a3b8; font-weight:bold;">
            Fases de aprendizaje estimadas: ${data.niveles_total || '1+'} niveles adaptativos.
        </div>
    `;

    // 2. Limpiar y dibujar el Consejo de GOnions dinámicos
    agentesContainer.innerHTML = '';
    
    // Safety check fallback to our mock data array
    const gonions = data.consejo_gonions || [];
    if (gonions.length === 0 && data.agentes) {
        // Fallback for our mock llm/client.go structure
        data.agentes.forEach(agente => {
            gonions.push({
                nombre: agente,
                familia: "Estándar",
                nivel: 1,
                personalidad: "En desarrollo",
                apariencia_sugerida: "Sprite genérico"
            });
        });
    }

    gonions.forEach(gonion => {
        const card = document.createElement('div');
        card.className = 'gonion-card';
        card.innerHTML = `
            <h3>🐹 ${gonion.nombre}</h3>
            <div class="meta">Familia: ${gonion.familia} | Rango Operativo: Nivel ${gonion.nivel}</div>
            <p class="desc"><strong>Estrategia:</strong> ${gonion.personalidad}</p>
            <p class="desc" style="margin-top:6px; color:#34d399;"><strong>Visual Asset:</strong> ${gonion.apariencia_sugerida}</p>
        `;
        agentesContainer.appendChild(card);
    });

    // NUEVO: Renderizar el Reto Inicial en el panel del Editor
    if (data.reto_actual) {
        document.getElementById('instrucciones-nivel').innerHTML = `
            <div style="background: #1e293b; padding: 10px; border-left: 3px solid #10b981; margin-bottom: 10px;">
                <strong>Mensaje Entrante:</strong> "${data.reto_actual.mensaje_gonion}"
            </div>
            <div style="color: #cbd5e1;"><strong>Misión Actual:</strong> ${data.reto_actual.instrucciones}</div>
        `;
        document.getElementById('codigo-editor').value = data.reto_actual.codigo_base;

        // Efecto UI: Enfocar automáticamente el panel del editor
        document.querySelector('.panel-editor').focus();
        logToTerminal('Estación de Código desbloqueada y lista para compilar.', 'success');
    }
}

// NUEVO: Evaluación de código
document.getElementById('btn-evaluar').addEventListener('click', () => {
    const codigo = document.getElementById('codigo-editor').value;
    const payload = {
        tipo: 'EVALUATE_CODE',
        nick: appState.nick,
        dominio: inputDominio.value.trim(),
        objetivo: inputObjetivo.value.trim(),
        nivel_actual: appState.nivelActual,
        codigo: codigo
    };
    logToTerminal('Enviando código al Consejo para evaluación...', 'system');
    ws.send(JSON.stringify(payload));
});
