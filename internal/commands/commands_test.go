package commands

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coccinella-labs/go-kit/internal/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "version")
	assert.NoError(t, err)
	assert.Contains(t, output, "0.1.0")
	assert.Contains(t, output, "kit version")
}

func TestVersionCommandOutput(t *testing.T) {
	rootCmd := NewRootCmd("1.2.3")

	output, err := executeCommand(rootCmd, "version")
	assert.NoError(t, err)
	assert.Contains(t, output, "1.2.3")
	assert.Contains(t, output, "kit version")
}

func TestAuthStatusNotLoggedIn(t *testing.T) {
	auth.Reset()
	tmpDir := t.TempDir()
	testStore := auth.NewSecureStore("kit-test-auth-status")
	testStore.SetFallbackDir(filepath.Join(tmpDir, ".kit"))
	testStore.SetForceFallback(true)
	auth.SetSecureStoreForTest(testStore)

	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "auth", "status")
	assert.NoError(t, err)
	assert.Contains(t, output, "github: not authenticated")

	auth.Reset()
	auth.SetSecureStoreForTest(auth.NewSecureStore("kit"))
}

func TestRootCommandHelp(t *testing.T) {
	rootCmd := NewRootCmd("0.1.0")
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "kit", rootCmd.Use)
}

func TestRootCommandSubcommands(t *testing.T) {
	rootCmd := NewRootCmd("0.1.0")

	expectedCommands := []string{"auth", "config", "github", "version"}

	for _, name := range expectedCommands {
		subCmd, _, _ := rootCmd.Find([]string{name})
		assert.NotNil(t, subCmd, "subcommand %s should exist", name)
	}
}

func TestAuthLoginMissingToken(t *testing.T) {
	auth.Reset()
	rootCmd := NewRootCmd("0.1.0")

	_, err := executeCommand(rootCmd, "auth", "login")
	assert.Error(t, err)
}

func TestAuthLoginWithToken(t *testing.T) {
	auth.Reset()
	tmpDir := t.TempDir()
	testStore := auth.NewSecureStore("kit-test-login")
	testStore.SetFallbackDir(filepath.Join(tmpDir, ".kit"))
	testStore.SetForceFallback(true)
	auth.SetSecureStoreForTest(testStore)

	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "auth", "login", "--token", "test-token")
	assert.NoError(t, err)
	assert.Contains(t, output, "Successfully authenticated")

	auth.Reset()
	auth.SetSecureStoreForTest(auth.NewSecureStore("kit"))
}

func TestConfigListEmpty(t *testing.T) {
	rootCmd := NewRootCmd("0.1.0")

	output, err := executeCommand(rootCmd, "config", "list")
	assert.NoError(t, err)
	assert.Contains(t, output, "api_url=https://api.github.com")
}

func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()

	return strings.TrimSpace(buf.String()), err
}