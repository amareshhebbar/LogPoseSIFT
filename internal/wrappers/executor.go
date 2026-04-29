package wrappers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// SafeExec runs a command with a strict timeout to prevent hung processes.
// Notice we do NOT pass a raw string to bash. We pass the explicit binary and arguments.
func SafeExec(binary string, args []string, timeoutMinutes time.Duration) (string, error) {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMinutes*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)

	// Capture standard output and standard error
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Check if the command was killed by our timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timed out after %d minutes", timeoutMinutes)
	}

	if err != nil {
		return "", fmt.Errorf("execution failed: %v | stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}