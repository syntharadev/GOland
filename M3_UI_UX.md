# Módulo 3: UI/UX - Estética Ciber-Zen, Kawaii y Componentes Interactivos

## 📌 1. Concepto de Diseño: Ciber-Zen & Kawaii
El lenguaje de diseño de **GOland** fusiona la sobriedad, legibilidad y el minimalismo de la filosofía **Ciber-Zen** (inspirada en terminales antiguas de computadoras, interfaces oscuras y de alta concentración) con detalles amigables, vibrantes y lúdicos del estilo **Kawaii** japonés (avatares redondos, paletas de colores dulces, micro-animaciones elásticas y retroalimentación lúdica).

El objetivo es crear una experiencia visual que mantenga el foco técnico absoluto del desarrollador y, a la vez, elimine la frustración del aprendizaje mediante recompensas estéticas que asombren a los jueces de hackatones a primera vista.

```
┌────────────────────────────────────────────────────────────────────────┐
│                          PALETA DE COLORES                             │
│  Abisal: #0F172A | Go-Cyan: #00ADD8 | Neon Pink: #FF2A85 | White: 80%  │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 🎨 2. Tokens de Estilo y Sistema de Diseño (CSS)

### Paleta de Colores
*   **Fondo Abisal (`--color-background`):** `#0F172A` (Azul muy oscuro y relajante).
*   **Acento Go-Cyan (`--color-primary`):** `#00ADD8` (El azul/celeste oficial de la marca Go, cargado de brillo digital).
*   **Acento Neon Pink (`--color-secondary`):** `#FF2A85` (Rosa neón para alertas, triggers Kawaii y estados desbloqueados).
*   **Cristal Translúcido (`--color-glass`):** `rgba(15, 23, 42, 0.4)` (Para paneles flotantes).
*   **Texto Ciber-Zen (`--color-text`):** `#E2E8F0` (Blanco grisáceo suave de alto contraste).

### Tipografía Asimétrica (Neo-Brutalismo Sutil)
Se implementa una jerarquía asimétrica para guiar visualmente la atención del desarrollador:
*   **Fondo/Estructura:** Números de nivel y códigos de módulo gigantescos translúcidos (ej. `opacity: 0.05` y `font-size: 8rem`) en el fondo de las tarjetas.
*   **Títulos y CTA:** Tipografía **Quicksand** o **Outfit** con bordes suaves y pesos redondos (`font-weight: 700`), aportando el toque Kawaii.
*   **Bloques de Código y Consolas:** Tipografía **Fira Code** o **JetBrains Mono** con ligaduras activas para mantener el rigor del entorno de desarrollo real.

---

## 🏗️ 3. Reglas de Estilo CSS para los Paneles "Hover Kawaii"

Los paneles interactivos de las misiones y de los GOmions utilizan un diseño translúcido hiper-premium que reacciona con una física de rebote elástica al interactuar con el cursor:

```css
/* Glassmorphism Base */
.glass-panel {
    background: rgba(15, 23, 42, 0.4);
    backdrop-filter: blur(24px) saturate(180%);
    -webkit-backdrop-filter: blur(24px) saturate(180%);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 24px;
    box-shadow: 0 4px 30px rgba(0, 0, 0, 0.15);
    color: #e2e8f0;
}

/* Efecto Hover Kawaii (Micro-interacción Elástica) */
.hover-kawaii {
    transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.hover-kawaii:hover {
    transform: translateY(-8px) scale(1.02);
    border: 1px solid rgba(0, 173, 216, 0.4); /* Resplandor Go-Cyan */
    box-shadow: 0 15px 35px rgba(0, 173, 216, 0.15);
    background: rgba(15, 23, 42, 0.6);
}
```

---

## ⛰️ 4. Parallax Parfait de 3 Capas
Para lograr que la landing page isométrica se sienta viva e interactiva al estilo de herramientas modernas como *Claude Code*, se implementa una técnica de paralaje de tres capas superpuestas en el DOM de la landing page:

```html
<div class="parallax-container" id="parallax-scene">
    <!-- Capa 1: Fondo (Cielo, estrellas y nubes lejanas) -->
    <div class="parallax-layer layer-background" data-depth="0.2"></div>
    
    <!-- Capa 2: Isla (Roca y templos estáticos de GOland) -->
    <div class="parallax-layer layer-island" data-depth="0.5">
        <img src="GOland.png" alt="Isla GOland">
    </div>
    
    <!-- Capa 3: Flotantes (Drones, nubes bajas y GOnions orbitando) -->
    <div class="parallax-layer layer-floaters" data-depth="0.8">
        <div class="gonion-bubble icon-mensajero"></div>
        <div class="gonion-bubble icon-profesor"></div>
        <div class="drone-scout"></div>
    </div>
</div>
```

### Animaciones @keyframes CSS Asociadas:
```css
/* Levitar de la Isla */
@keyframes levitateIsland {
    0% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
    100% { transform: translateY(0px); }
}

.layer-island img {
    animation: levitateIsland 6s ease-in-out infinite;
}

/* Movimiento de las Nubes del Fondo */
@keyframes driftClouds {
    0% { background-position: 0 0; }
    100% { background-position: 1000px 0; }
}

.layer-background {
    background-image: url('nubes-pixel.png');
    animation: driftClouds 60s linear infinite;
}

/* Drones y GOnions Flotando (Ritmo asincrónico para profundidad) */
@keyframes bobFloaters {
    0% { transform: translateY(0px) rotate(0deg); }
    50% { transform: translateY(-15px) rotate(2deg); }
    100% { transform: translateY(0px) rotate(0deg); }
}

.layer-floaters .gonion-bubble {
    animation: bobFloaters 4.5s ease-in-out infinite;
}
```

---

## 💬 5. Componente: Chat Holográfico Arrastrable

El punto focal de interacción con los agentes de IA es una consola holográfica flotante de cristal que puede reposicionarse dinámicamente por la interfaz del usuario:

```
┌────────────────────────────────────────────────────────┐
│  [ • ]  [ 🗕 ]  [ 🗖 ]  CONSOLA GOMION - PROFESOR      │  <-- Drag Handle
├────────────────────────────────────────────────────────┤
│                                                        │
│  [Profesor]: Analizando tu bucle concurrente...        │
│  █                                                     │
└────────────────────────────────────────────────────────┘
```

### Especificaciones Técnicas del Componente
1.  **Modo Escritorio (Draggable Window):**
    *   Contenedor fijo (`position: fixed`) con alto índice Z (`z-index: 1000`).
    *   Arrastrable mediante interacción JS capturando eventos `mousedown`, `mousemove` y `mouseup` enlazados exclusivamente al header superior del chat (`.drag-handle`).
2.  **Modo Móvil (Bottom Sheet Fallback):**
    *   A través de media queries (`@media (max-width: 768px)`), el posicionamiento arrastrable se deshabilita por completo.
    *   El contenedor se ancla a la base de la ventana de la pantalla (`bottom: 0; left: 0; width: 100%`), comportándose como una tarjeta desplegable elástica (Bottom Sheet) que se maximiza deslizando el dedo hacia arriba.
3.  **Estados del Chat:**
    *   **Minimizado:** Se colapsa en un widget circular Kawaii animado en la esquina inferior derecha con la cara del GOnion activo.
    *   **Maximizado:** Se despliega el panel holográfico de cristal completo con el editor de historial y prompt.

---

## ⚡ 6. Efectos de Terminal Retro / Ciber-Zen

### Text Scramble (Cargador de Terminal)
Cuando un GOnion está procesando una llamada de RAG o evaluando el código del usuario, el cuadro de respuesta del chat flotante muestra una secuencia en bucle de caracteres aleatorios que mutan velozmente hasta revelar el veredicto final:

```javascript
// Simulación conceptual de Text Scramble
const chars = '!<>-_\\/[]{}—=+*^?#________';
function scrambleText(targetElement, finalString) {
    let iteration = 0;
    const interval = setInterval(() => {
        targetElement.innerText = finalString.split("")
            .map((char, index) => {
                if(index < iteration) return finalString[index];
                return chars[Math.floor(Math.random() * chars.length)];
            }).join("");
        
        if(iteration >= finalString.length) clearInterval(interval);
        iteration += 1/3;
    }, 30);
}
```

### Cursor Ciber-Zen Parpadeante (█)
El prompt y las salidas de texto de los GOmions finalizan obligatoriamente con el bloque de terminal retro de Go parpadeante, lo que inyecta la inmersión de interactuar con el sistema de control integrado en la nave espacial de GOland:

```css
.terminal-cursor::after {
    content: "█";
    color: var(--color-primary); /* Go-Cyan */
    animation: blinkCursor 0.8s steps(2, start) infinite;
}

@keyframes blinkCursor {
    to { visibility: hidden; }
}
```
