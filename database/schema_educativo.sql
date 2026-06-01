-- ==========================================================================
-- 🧬 SCHEMA DE INFRAESTRUCTURA EDUCATIVA Y TEMARIO GAMIFICADO DE GOLAND
-- Compatible con PostgreSQL y Supabase SQL Editor
-- ==========================================================================

-- 1. TABLA: niveles_temario (El Plan de Estudios / Roadmap de Go)
CREATE TABLE IF NOT EXISTS niveles_temario (
    nivel_id INT PRIMARY KEY,
    titulo_leccion VARCHAR(255) NOT NULL,
    descripcion TEXT NOT NULL,
    experiencia_requerida INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

-- 2. TABLA: go_documentacion (Índice de Consulta Rápida / RAG Local)
CREATE TABLE IF NOT EXISTS go_documentacion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tema VARCHAR(100) UNIQUE NOT NULL,
    resumen TEXT NOT NULL,
    url_oficial TEXT NOT NULL,
    nivel_requerido INT REFERENCES niveles_temario(nivel_id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

-- 3. TABLA: retos_codigo (Los Desafíos del Estudiante Auditados por GOnions)
CREATE TABLE IF NOT EXISTS retos_codigo (
    reto_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nivel_id INT REFERENCES niveles_temario(nivel_id) ON DELETE CASCADE,
    codigo_base TEXT NOT NULL,
    test_validado TEXT NOT NULL, -- Expresión regular o cadena de salida esperada (expected output)
    gonion_evaluador VARCHAR(100) DEFAULT 'Profesor', -- Ej: "Profesor", "Guardiana", "Cronometradora"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

-- ==========================================================================
-- 🛡️ CONFIGURACIÓN DE SEGURIDAD A NIVEL DE FILA (RLS)
-- Supabase exige políticas claras para resguardar o liberar datos
-- ==========================================================================

ALTER TABLE niveles_temario ENABLE ROW LEVEL SECURITY;
ALTER TABLE go_documentacion ENABLE ROW LEVEL SECURITY;
ALTER TABLE retos_codigo ENABLE ROW LEVEL SECURITY;

-- Políticas de lectura pública (cualquier usuario registrado o anónimo puede leer el contenido educativo)
CREATE POLICY "Permitir lectura de niveles a todos" 
    ON niveles_temario FOR SELECT USING (true);

CREATE POLICY "Permitir lectura de documentacion a todos" 
    ON go_documentacion FOR SELECT USING (true);

CREATE POLICY "Permitir lectura de retos de codigo a todos" 
    ON retos_codigo FOR SELECT USING (true);

-- Políticas de escritura restrictiva (sólo administradores de Supabase / Service Role o backend con API Key válida)
CREATE POLICY "Restringir inserciones a roles autorizados en niveles" 
    ON niveles_temario FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Restringir inserciones a roles autorizados en docs" 
    ON go_documentacion FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Restringir inserciones a roles autorizados en retos" 
    ON retos_codigo FOR ALL USING (auth.role() = 'authenticated');

-- ==========================================================================
-- 🌱 INSERCIÓN DE REGISTROS DE INICIALIZACIÓN (MOCK / SEMILLAS INICIALES)
-- ==========================================================================

-- Semillas para niveles_temario
INSERT INTO niveles_temario (nivel_id, titulo_leccion, descripcion, experiencia_requerida) VALUES
(1, 'Variables y Asignación Corta', 'Aprende los fundamentos de tipos estáticos, var y el operador cuántico de inferencia corta := en Go.', 0),
(2, 'Estructuras de Control y Control Flow', 'Domina if/else, los bucles for únicos de Go y la estructura de decisión switch para orquestar la simulación.', 100),
(3, 'Estructuras y Punteros', 'Entiende cómo agrupar datos de forma limpia con structs y manejar referencias en memoria con punteros sin riesgo.', 300),
(5, 'Paquetes, Imports y Interfaces', 'Aprende a modularizar tus aplicaciones Go con paquetes personalizados e inyecta polimorfismo mediante interfaces dinámicas.', 800),
(10, 'Concurrencia, Goroutines y Canales', 'Desbloquea el nivel definitivo: ejecuta miles de tareas concurrentes ultraligeras y comunícalas mediante canales de sincronía.', 2000)
ON CONFLICT (nivel_id) DO UPDATE 
SET titulo_leccion = EXCLUDED.titulo_leccion, descripcion = EXCLUDED.descripcion;

-- Semillas para go_documentacion
INSERT INTO go_documentacion (tema, resumen, url_oficial, nivel_requerido) VALUES
('Asignación Corta (:=)', 'Declara e inicializa variables locales con inferencia de tipo automática de manera eficiente y ágil.', 'https://go.dev/tour/basics/10', 1),
('Sintaxis de Bucles For', 'Go solo cuenta con una palabra clave para bucles: "for". Sirve para bucles clásicos, condicionales y colecciones iterables (range).', 'https://go.dev/tour/flowcontrol/1', 2),
('Punteros (*T y &)', 'Mapea la dirección de memoria de un dato directamente para eficiencia extrema en llamadas de funciones.', 'https://go.dev/tour/moretypes/1', 3),
('Goroutines', 'Un hilo cuántico de ejecución extremadamente ligero administrado nativamente por el runtime de Go.', 'https://go.dev/tour/concurrency/1', 10)
ON CONFLICT (tema) DO UPDATE 
SET resumen = EXCLUDED.resumen, url_oficial = EXCLUDED.url_oficial, nivel_requerido = EXCLUDED.nivel_requerido;

-- Semillas para retos_codigo
INSERT INTO retos_codigo (nivel_id, codigo_base, test_validado, gonion_evaluador) VALUES
(
 1, 
 'package main\n\nimport "fmt"\n\nfunc main() {\n    // Declara mensaje de tipo string con el valor "Hola GOland"\n    mensaje := "Hola GOland"\n    fmt.Println(mensaje)\n}', 
 'Hola GOland', 
 'Profesor'
),
(
 2, 
 'package main\n\nimport "fmt"\n\nfunc main() {\n    // Haz un bucle de 1 a 3 imprimiendo el número en pantalla\n    for i := 1; i <= 3; i++ {\n        fmt.Println(i)\n    }\n}', 
 '1\n2\n3', 
 'Bibliotecaria'
),
(
 10, 
 'package main\n\nimport "fmt"\n\nfunc worker(ch chan string) {\n    ch <- "Goroutine Completa"\n}\n\nfunc main() {\n    ch := make(chan string)\n    go worker(ch)\n    fmt.Println(<-ch)\n}', 
 'Goroutine Completa', 
 'Cronometradora'
);
