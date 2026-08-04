package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
)

const modulePath = "github.com/m-tkg/md2site"

// Version reports the module version stamped by "go install
// <module>@<version>"; source builds report "devel".
func Version() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "devel"
}

// runUpgrade reinstalls the latest release. md2site is distributed via
// "go install", so the Go toolchain is guaranteed to exist on any machine
// that installed it — delegating back to it is the simplest reliable
// self-update.
func runUpgrade() error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go command not found; install Go and run: go install %s@latest", modulePath)
	}
	fmt.Printf("md2site %s -> go install %s@latest\n", Version(), modulePath)
	cmd := exec.Command(gobin, "install", modulePath+"@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}
	if out, err := exec.Command(gobin, "version", "-m", installedBinary(gobin)).Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "\tmod\t") {
				fmt.Println("installed:", strings.TrimSpace(line))
			}
		}
	}
	fmt.Println("upgrade done")
	return nil
}

// installedBinary returns the path "go install" wrote to, following the
// GOBIN > GOPATH/bin resolution go itself uses.
func installedBinary(gobin string) string {
	dir := ""
	if out, err := exec.Command(gobin, "env", "GOBIN").Output(); err == nil {
		dir = trimNL(string(out))
	}
	if dir == "" {
		if out, err := exec.Command(gobin, "env", "GOPATH").Output(); err == nil {
			dir = trimNL(string(out)) + "/bin"
		}
	}
	return dir + "/md2site"
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
