package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Generate(args []string) {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate failed: %v\n", err)
		os.Exit(1)
	}

	if err := generateKitProtos(root); err != nil {
		fmt.Fprintf(os.Stderr, "generate kit protos failed: %v\n", err)
		os.Exit(1)
	}

	if err := generateRootProtos(root, "platform", "platform"); err != nil {
		fmt.Fprintf(os.Stderr, "generate platform protos failed: %v\n", err)
		os.Exit(1)
	}

	if err := generateRootProtos(root, "services", "services"); err != nil {
		fmt.Fprintf(os.Stderr, "generate service protos failed: %v\n", err)
		os.Exit(1)
	}

	if err := generateServiceTSSDKs(root); err != nil {
		fmt.Fprintf(os.Stderr, "generate ts sdk failed: %v\n", err)
		os.Exit(1)
	}

	if err := runSqlc(root); err != nil {
		fmt.Fprintf(os.Stderr, "sqlc generate failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("generate complete")
}

func goPlugins(out string) []map[string]any {
	return []map[string]any{
		{"plugin": "go", "out": out, "opt": []string{"paths=source_relative"}},
		{"plugin": "go-grpc", "out": out, "opt": []string{"paths=source_relative"}},
		{"plugin": "grpc-gateway", "out": out, "opt": []string{"paths=source_relative", "generate_unbound_methods=true"}},
	}
}

func generateKitProtos(root string) error {
	template := map[string]any{
		"version": "v1",
		"plugins": goPlugins("kit/kit-go"),
	}
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	if err := runBuf(root, "generate", "--path", "proto/plantx/kit", "--template", string(templateBytes)); err != nil {
		return err
	}
	if err := reorganizeKitProtoOutputs(root); err != nil {
		return err
	}
	return nil
}

func generateRootProtos(root, rootName, out string) error {
	dirs, err := protoDirs(root, rootName)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}
	template := map[string]any{
		"version": "v1",
		"plugins": goPlugins(out),
	}
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	args := []string{"generate"}
	for _, d := range dirs {
		args = append(args, "--path", d)
	}
	args = append(args, "--template", string(templateBytes))
	return runBuf(root, args...)
}

func protoDirs(root, rootName string) ([]string, error) {
	var dirs []string
	rootDir := filepath.Join(root, rootName)
	if err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".proto") {
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(root, dir)
			if err != nil {
				return err
			}
			dirs = append(dirs, filepath.ToSlash(rel))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// De-duplicate while preserving order.
	seen := make(map[string]struct{})
	var out []string
	for _, d := range dirs {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out, nil
}

func reorganizeKitProtoOutputs(root string) error {
	moves := []struct {
		src string
		dst string
	}{
		{filepath.Join(root, "kit", "kit-go", "plantx", "kit", "authz.pb.go"), filepath.Join(root, "kit", "kit-go", "proto", "authz", "authz.pb.go")},
		{filepath.Join(root, "kit", "kit-go", "plantx", "kit", "context.pb.go"), filepath.Join(root, "kit", "kit-go", "proto", "context", "context.pb.go")},
		{filepath.Join(root, "kit", "kit-go", "plantx", "kit", "event.pb.go"), filepath.Join(root, "kit", "kit-go", "proto", "event", "event.pb.go")},
	}
	for _, m := range moves {
		if _, err := os.Stat(m.src); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := os.MkdirAll(filepath.Dir(m.dst), 0755); err != nil {
			return err
		}
		if err := os.Rename(m.src, m.dst); err != nil {
			return err
		}
	}
	// Remove generated transient directories that are not part of the source tree.
	for _, d := range []string{
		filepath.Join(root, "kit", "kit-go", "plantx"),
		filepath.Join(root, "kit", "kit-go", "google"),
	} {
		_ = os.RemoveAll(d)
	}
	return nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "buf.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("project root not found (missing go.work or buf.yaml)")
}

func generateServiceTSSDKs(root string) error {
	servicesDir := filepath.Join(root, "services")
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		protoDir := filepath.Join("services", name, "api")
		if _, err := os.Stat(filepath.Join(root, protoDir, name+".proto")); err != nil {
			continue
		}
		outDir := filepath.Join("services", name, "web", name+"-sdk-api", "src", "generated")
		if err := os.MkdirAll(filepath.Join(root, outDir), 0755); err != nil {
			return fmt.Errorf("service %s: %w", name, err)
		}
		template := map[string]any{
			"version": "v1",
			"plugins": []map[string]any{
				{
					"plugin": "ts",
					"out":    outDir,
					"opt":    []string{"generate_dependencies", "optimize_code_size"},
				},
			},
		}
		templateBytes, err := json.Marshal(template)
		if err != nil {
			return fmt.Errorf("service %s: %w", name, err)
		}
		if err := runBuf(root, "generate", "--path", protoDir, "--template", string(templateBytes)); err != nil {
			return fmt.Errorf("service %s: %w", name, err)
		}
	}
	return nil
}

func runSqlc(root string) error {
	cmd := exec.Command("sqlc", "generate")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runBuf(root string, args ...string) error {
	buf, err := lookBuf(root)
	if err != nil {
		return err
	}
	cmd := exec.Command(buf, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = withNodeBinPath(os.Environ(), root)
	return cmd.Run()
}

func lookBuf(root string) (string, error) {
	if p, err := exec.LookPath("buf"); err == nil {
		return p, nil
	}
	candidates := []string{
		filepath.Join(root, "node_modules", ".bin", "buf"),
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, filepath.Join(root, "node_modules", ".bin", "buf.CMD"))
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("buf not found in PATH; install github.com/bufbuild/buf or run `pnpm add -D @bufbuild/buf`")
}

func withNodeBinPath(env []string, root string) []string {
	absBin, err := filepath.Abs(filepath.Join(root, "node_modules", ".bin"))
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
