package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// JSONRPCRequest representa una petición JSON-RPC 2.0
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse representa una respuesta JSON-RPC 2.0
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError representa un error en la respuesta JSON-RPC
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPClient encapsula la conexión vía stdio a un servidor MCP
type MCPClient struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	scanner   *bufio.Scanner
	idGen     int64
	pending   map[int64]chan *JSONRPCResponse
	pendingMu sync.Mutex
	traceFile *os.File
	closeChan chan struct{}
	wg        sync.WaitGroup
}

// GlobalMCPClients mantiene registros de clientes activos para poder cerrarlos ordenadamente
var (
	ActiveClients   = make(map[string]*MCPClient)
	ActiveClientsMu sync.Mutex
)

// IniciarMCPClient arranca un servidor MCP como subproceso y establece la comunicación stdio
func IniciarMCPClient(clientName string, command string, args ...string) (*MCPClient, error) {
	log.Printf("MCP Client [%s]: Iniciando servidor con comando: %s %v", clientName, command, args)

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error al obtener stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error al obtener stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error al obtener stderr pipe: %w", err)
	}

	// Abrir archivo de trazas para MiroShark
	traceFile, err := os.OpenFile("mcp_trace.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Aviso: No se pudo abrir mcp_trace.log para tracing: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error al iniciar subproceso MCP: %w", err)
	}

	client := &MCPClient{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		scanner:   bufio.NewScanner(stdout),
		pending:   make(map[int64]chan *JSONRPCResponse),
		traceFile: traceFile,
		closeChan: make(chan struct{}),
	}

	// Escuchar stderr del subproceso en segundo plano
	client.wg.Add(1)
	go func() {
		defer client.wg.Done()
		r := bufio.NewReader(stderr)
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			log.Printf("MCP Server [%s] STDERR: %s", clientName, line)
		}
	}()

	// Escuchar respuestas de stdout en segundo plano
	client.wg.Add(1)
	go client.listenStdout(clientName)

	// Inicializar protocolo MCP
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Initialize(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("error al inicializar protocolo MCP: %w", err)
	}

	ActiveClientsMu.Lock()
	ActiveClients[clientName] = client
	ActiveClientsMu.Unlock()

	return client, nil
}

// listenStdout lee línea por línea el stdout del servidor y despacha a canales pendientes
func (m *MCPClient) listenStdout(clientName string) {
	defer m.wg.Done()
	for m.scanner.Scan() {
		line := m.scanner.Bytes()

		// Registrar traza de entrada
		m.writeTrace("IN", line)

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			log.Printf("MCP Client [%s]: Error parseando respuesta JSON-RPC: %v. Line: %s", clientName, err, string(line))
			continue
		}

		if resp.ID != nil {
			m.pendingMu.Lock()
			ch, exists := m.pending[*resp.ID]
			if exists {
				ch <- &resp
				delete(m.pending, *resp.ID)
			}
			m.pendingMu.Unlock()
		}
	}
	if err := m.scanner.Err(); err != nil {
		log.Printf("MCP Client [%s]: Error leyendo stdout scanner: %v", clientName, err)
	}
}

// writeTrace escribe la traza JSON-RPC a mcp_trace.log e imprime a la consola
func (m *MCPClient) writeTrace(direction string, data []byte) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	traceLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, direction, string(data))

	log.Printf("MCP Trace: %s", traceLine)

	if m.traceFile != nil {
		_, _ = m.traceFile.WriteString(traceLine)
	}
}

// sendRequest envía una petición formateada por stdin y espera su respuesta en segundo plano
func (m *MCPClient) sendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	id := atomic.AddInt64(&m.idGen, 1)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error serializando petición JSON-RPC: %w", err)
	}

	// Canal para recibir la respuesta
	ch := make(chan *JSONRPCResponse, 1)

	m.pendingMu.Lock()
	m.pending[id] = ch
	m.pendingMu.Unlock()

	// Registrar traza de salida
	m.writeTrace("OUT", data)

	// Escribir al stdin con nueva línea
	_, err = m.stdin.Write(append(data, '\n'))
	if err != nil {
		m.pendingMu.Lock()
		delete(m.pending, id)
		m.pendingMu.Unlock()
		return nil, fmt.Errorf("error escribiendo a stdin de MCP: %w", err)
	}

	// Esperar respuesta o cancelación de contexto
	select {
	case <-ctx.Done():
		m.pendingMu.Lock()
		delete(m.pending, id)
		m.pendingMu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		return resp, nil
	}
}

// Initialize envía la petición 'initialize' requerida por el protocolo MCP
func (m *MCPClient) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "GOland-Orchestrator",
			"version": "1.0.0",
		},
	}

	resp, err := m.sendRequest(ctx, "initialize", params)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("error del servidor en inicialización: %s (código: %d)", resp.Error.Message, resp.Error.Code)
	}

	// Enviar notificación 'initialized' requerida por la especificación MCP
	initializedNotif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	notifData, err := json.Marshal(initializedNotif)
	if err == nil {
		m.writeTrace("OUT", notifData)
		_, _ = m.stdin.Write(append(notifData, '\n'))
	}

	return nil
}

// CallTool ejecuta una herramienta remota en el servidor MCP y devuelve la respuesta
func (m *MCPClient) CallTool(ctx context.Context, toolName string, arguments interface{}) (string, error) {
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	resp, err := m.sendRequest(ctx, "tools/call", params)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("error ejecutando herramienta %s: %s (código: %d)", toolName, resp.Error.Message, resp.Error.Code)
	}

	return string(resp.Result), nil
}

// Close detiene el subproceso del servidor de forma limpia
func (m *MCPClient) Close() {
	close(m.closeChan)
	_ = m.stdin.Close()
	_ = m.stdout.Close()
	_ = m.stderr.Close()

	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
	}

	if m.traceFile != nil {
		_ = m.traceFile.Close()
	}

	m.wg.Wait()
	log.Println("MCP Client cerrado correctamente.")
}
