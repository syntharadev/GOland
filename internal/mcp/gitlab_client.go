package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PipelineResult representa el veredicto del pipeline de GitLab CI
type PipelineResult struct {
	Status    string `json:"status"` // "success" o "failed"
	Log       string `json:"log"`    // Salida detallada del compilador
	JobURL    string `json:"job_url"`
	Duration  string `json:"duration"`
}

// EjecutarPipelineGo simula la ejecución de un pipeline de CI/CD en GitLab que corre go build/test
func EjecutarPipelineGo(codigoEstudiante string) (string, error) {
	codigo := strings.TrimSpace(codigoEstudiante)

	var res PipelineResult
	res.JobURL = "https://gitlab.com/goval-ci/jobs/92847162"
	res.Duration = "4.2s"

	// Validaciones básicas de compilador de Go simuladas
	if codigo == "" {
		res.Status = "failed"
		res.Log = "GitLab CI Runner [Go Builder]:\n# compilation failed\n/tmp/build/main.go:1:1: empty source file"
	} else if !strings.Contains(codigo, "package main") {
		res.Status = "failed"
		res.Log = "GitLab CI Runner [Go Builder]:\n# compilation failed\n/tmp/build/main.go:1:1: expected 'package', found 'EOF'"
	} else if !strings.Contains(codigo, "func main()") {
		res.Status = "failed"
		res.Log = "GitLab CI Runner [Go Builder]:\n# compilation failed\n/tmp/build/main.go: missing main function in package main"
	} else if strings.Contains(codigo, "fmt.Println") && !strings.Contains(codigo, "import \"fmt\"") && !strings.Contains(codigo, "import (") {
		res.Status = "failed"
		res.Log = "GitLab CI Runner [Go Builder]:\n# compilation failed\n/tmp/build/main.go: undefined: fmt"
	} else if strings.Count(codigo, "{") != strings.Count(codigo, "}") {
		res.Status = "failed"
		res.Log = "GitLab CI Runner [Go Builder]:\n# compilation failed\n/tmp/build/main.go: syntax error: unexpected semicolon or newline, expecting }"
	} else {
		res.Status = "success"
		res.Log = "GitLab CI Runner [Go Builder]:\n$ go build -o app ./...\n$ go test -v ./...\n=== RUN   TestMision\n--- PASS: TestMision (0.01s)\nPASS\nok  \tgoval-ci/build\t0.015s\nJob succeeded: Build and tests passed."
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		return "", fmt.Errorf("error serializando resultado del pipeline GitLab: %w", err)
	}

	return string(jsonData), nil
}
