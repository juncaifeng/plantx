// Package scaffold provides service scaffolding helpers.
package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ServiceOptions controls scaffold output.
type ServiceOptions struct {
	WithUI      bool
	WithGateway bool
}

// Service creates a new service directory tree.
func Service(name string, opts ServiceOptions) error {
	if name == "" {
		return fmt.Errorf("service name is required")
	}

	serviceDir := filepath.Join("services", name)
	dirs := []string{
		filepath.Join(serviceDir, "api"),
		filepath.Join(serviceDir, "internal", "domain"),
		filepath.Join(serviceDir, "internal", "app"),
		filepath.Join(serviceDir, "internal", "infra", "sqlc"),
		filepath.Join(serviceDir, "internal", "infra", "repo"),
		filepath.Join(serviceDir, "internal", "interfaces", "grpc"),
		filepath.Join(serviceDir, "migrations"),
		filepath.Join(serviceDir, "cmd"),
	}
	if opts.WithUI {
		dirs = append(dirs,
			filepath.Join(serviceDir, "web", name+"-sdk-api", "src", "generated"),
			filepath.Join(serviceDir, "web", name+"-ui", "src"),
		)
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		filepath.Join(serviceDir, "go.mod"):                                       serviceGoMod(name),
		filepath.Join(serviceDir, "api", name+".proto"):                           serviceProto(name),
		filepath.Join(serviceDir, "sqlc.yaml"):                                    sqlcYaml(),
		filepath.Join(serviceDir, "internal", "domain", name+".go"):               domainStub(name),
		filepath.Join(serviceDir, "internal", "app", "service.go"):                appStub(name),
		filepath.Join(serviceDir, "internal", "interfaces", "grpc", "handler.go"): handlerStub(name),
		filepath.Join(serviceDir, "cmd", "main.go"):                               mainStub(name),
		filepath.Join(serviceDir, "Dockerfile"):                                   dockerfileStub(name),
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	if err := generateProto(serviceDir); err != nil {
		return fmt.Errorf("scaffold generated, but failed to generate proto code: %w", err)
	}

	return nil
}

func generateProto(serviceDir string) error {
	buf, err := lookBuf()
	if err != nil {
		return err
	}
	template := map[string]any{
		"version": "v1",
		"plugins": []map[string]any{
			{"plugin": "go", "out": "services", "opt": []string{"paths=source_relative"}},
			{"plugin": "go-grpc", "out": "services", "opt": []string{"paths=source_relative"}},
			{"plugin": "grpc-gateway", "out": "services", "opt": []string{"paths=source_relative", "generate_unbound_methods=true"}},
		},
	}
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	cmd := exec.Command(buf, "generate", "--path", filepath.Join(serviceDir, "api"), "--template", string(templateBytes))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = withNodeBinPath(os.Environ())
	return cmd.Run()
}

func lookBuf() (string, error) {
	if p, err := exec.LookPath("buf"); err == nil {
		return p, nil
	}
	// Try local node_modules fallback.
	candidates := []string{
		filepath.Join("node_modules", ".bin", "buf"),
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, filepath.Join("node_modules", ".bin", "buf.CMD"))
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("buf not found in PATH; install github.com/bufbuild/buf or run `pnpm add -D @bufbuild/buf`")
}

func withNodeBinPath(env []string) []string {
	absBin, err := filepath.Abs(filepath.Join("node_modules", ".bin"))
	if err != nil {
		return env
	}
	for i, e := range env {
		if len(e) > 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + absBin + string(os.PathListSeparator) + e[5:]
			return env
		}
	}
	return append(env, "PATH="+absBin)
}

func serviceGoMod(name string) string {
	return fmt.Sprintf(`module github.com/plantx/services/%s

go 1.22

require github.com/plantx/kit/kit-go v0.0.0

replace github.com/plantx/kit/kit-go => ../../kit/kit-go
`, name)
}

func serviceProto(name string) string {
	return fmt.Sprintf(`syntax = "proto3";

package plantx.%s.v1;

option go_package = "github.com/plantx/services/%s/api";

service %sService {
  rpc Ping(PingRequest) returns (PongResponse);
}

message PingRequest {
  string message = 1;
}

message PongResponse {
  string message = 1;
}
`, name, name, title(name))
}

func sqlcYaml() string {
	return `version: "2"
sql:
  - schema: "migrations"
    queries: "internal/infra/sqlc"
    engine: "postgresql"
    gen:
      go:
        package: "sqlc"
        out: "internal/infra/sqlc"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
`
}

func domainStub(name string) string {
	return fmt.Sprintf(`package domain

// %s represents the aggregate root.
type %s struct {
	ID string
}
`, title(name), title(name))
}

func appStub(_ string) string {
	return `package app

import "context"

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Ping(ctx context.Context, msg string) string {
	return msg
}
`
}

func handlerStub(name string) string {
	return fmt.Sprintf(`package grpc

import (
	"context"

	pb "github.com/plantx/services/%s/api"
	"github.com/plantx/services/%s/internal/app"
)

type Handler struct {
	app *app.Service
}

func NewHandler(app *app.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	return &pb.PongResponse{Message: h.app.Ping(ctx, req.Message)}, nil
}
`, name, name)
}

func mainStub(name string) string {
	return fmt.Sprintf(`package main

import (
	"context"
	"fmt"

	"github.com/plantx/kit/kit-go/server"
)

func main() {
	srv := server.New(server.Options{GRPCPort: 8080})
	fmt.Printf("starting %%s service on :8080\n", "%s")
	if err := srv.Run(context.Background()); err != nil {
		panic(err)
	}
}
`, name)
}

func dockerfileStub(name string) string {
	return fmt.Sprintf(`FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /bin/%s ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/%s .
CMD ["./%s"]
`, name, name, name)
}

func title(name string) string {
	if len(name) == 0 {
		return name
	}
	return string(name[0]-32) + name[1:]
}
