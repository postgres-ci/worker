package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

func CheckoutToRevision(repository, workingDir, branch, revision string) (string, error) {

	path := filepath.Join(workingDir, branch, revision)
	{

		cmd := exec.Command("git", "clone", "-q", "--single-branch", "--branch", branch, repository, path)
		cmd.Stderr = &bytes.Buffer{}

		if err := cmd.Run(); err != nil {

			return "", fmt.Errorf("git: %s", cmd.Stderr.(*bytes.Buffer))
		}
	}

	cmd := exec.Command("git", "checkout", "-q", revision)
	cmd.Dir = path
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {

		return "", fmt.Errorf("git: %s", cmd.Stderr.(*bytes.Buffer))
	}

	return path, nil
}
