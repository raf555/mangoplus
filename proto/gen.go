//go:build ignore

// Regenerates Go bindings for all .proto files in this directory into
// ../internal/proto. Run with `go run proto/gen.go` from the repo root.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

const (
	protoDir  = "proto"
	outDir    = "internal/proto"
	goPackage = "github.com/raf555/mangoplus/internal/proto"
)

func main() {
	matches, err := filepath.Glob(filepath.Join(protoDir, "*.proto"))
	if err != nil {
		fail(err)
	}
	if len(matches) == 0 {
		fail(fmt.Errorf("no .proto files found in %s", protoDir))
	}
	slices.Sort(matches)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail(err)
	}

	args := []string{
		"--go_out=" + outDir,
		"--go_opt=paths=source_relative",
		"--proto_path=" + protoDir,
	}
	for _, p := range matches {
		args = append(args, "--go_opt=M"+filepath.Base(p)+"="+goPackage)
	}
	args = append(args, matches...)

	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen:", err)
	os.Exit(1)
}
