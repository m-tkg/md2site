// Package server provides the serve subcommand: build into a temporary
// directory and serve it over HTTP for preview.
package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/m-tkg/md2site/internal/build"
)

func BuildAndServe(cfg build.Config, port int) error {
	tmp, err := os.MkdirTemp("", "md2site-serve-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cfg.OutputDir = tmp
	if err := build.Run(cfg); err != nil {
		return err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("md2site: serving on http://%s/ (Ctrl-C to stop)\n", addr)
	return http.ListenAndServe(addr, http.FileServer(http.Dir(tmp)))
}
