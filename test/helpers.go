package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ChangeToRootDir() error {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	root := filepath.Clean(strings.TrimSpace(string(out)))
	if err := os.Chdir(root); err != nil {
		return err
	}
	return nil
}
