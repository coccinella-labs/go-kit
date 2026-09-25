package commands

import (
	"testing"

	"github.com/coccinella-labs/go-kit/internal/auth"
	"github.com/coccinella-labs/go-kit/internal/exitcode"
	"github.com/stretchr/testify/assert"
)

func TestDoctorCommand(t *testing.T) {
	auth.Reset()
	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "doctor")
	if err != nil {
		assert.Contains(t, output, "Configuration")
		assert.Contains(t, output, "Authentication")
		assert.Contains(t, output, "Git")
		assert.Contains(t, output, "Network")
		assert.Contains(t, output, "issue(s) found")
	} else {
		assert.Contains(t, output, "kit is ready")
	}
}

func TestDoctorExitCode(t *testing.T) {
	auth.Reset()
	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "doctor")
	if err != nil {
		assert.True(t, exitcode.Is(err))
		assert.Equal(t, exitcode.ErrGeneral, exitcode.Code(err))
		assert.Contains(t, output, "issue(s) found")
	} else {
		assert.Contains(t, output, "kit is ready")
	}
}
