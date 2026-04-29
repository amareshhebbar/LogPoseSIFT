package wrappers

import (
	"fmt"
)

// GetWindowsInfo securely executes Volatility 3's windows.info plugin.
// It takes the path to a memory dump and returns the terminal output.
func GetWindowsInfo(memoryDumpPath string) (string, error) {
	// Inside the SIFT VM, Volatility 3 is called via 'vol'
	binary := "vol"
	
	// We strictly define the arguments. The AI only provides the path.
	args := []string{
		"-f", memoryDumpPath,
		"windows.info",
	}

	// Execute with a 2-minute timeout
	output, err := SafeExec(binary, args, 2)
	if err != nil {
		return "", fmt.Errorf("volatility windows.info failed: %w", err)
	}

	return output, nil
}