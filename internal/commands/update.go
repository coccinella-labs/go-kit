package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/coccinella-labs/go-kit/internal/exitcode"
	"github.com/spf13/cobra"
)

const (
	githubAPIURL = "https://api.github.com/repos/coccinella-labs/go-kit/releases/latest"
)

type InstallMethod string

const (
	InstallMethodUnknown  InstallMethod = "unknown"
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodGo       InstallMethod = "go"
	InstallMethodScript   InstallMethod = "script"
	InstallMethodManual   InstallMethod = "manual"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for kit updates",
		Long:  `Check for the latest kit release and suggest an update path.`,
		RunE:  runUpdate,
	}

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	currentVersion := cmd.Root().Version
	if currentVersion == "" {
		currentVersion = "dev"
	}

	info(cmd, "Checking for updates...")
	latestVersion, err := getLatestVersion(cmd.Context())
	if err != nil {
		if isNetworkError(err) {
			return exitcode.New(exitcode.ErrNetwork, fmt.Errorf("failed to check for updates: %w", err))
		}
		return exitcode.New(exitcode.ErrAPI, fmt.Errorf("failed to check for updates: %w", err))
	}

	if latestVersion == currentVersion {
		info(cmd, "You are running the latest version ("+currentVersion+")")
		return nil
	}

	warn(cmd, "A new version is available: "+latestVersion+" (current: "+currentVersion+")")
	cmd.Println()

	method := detectInstallMethod()
	cmd.Printf("Installed via: %s\n", method)
	cmd.Println()
	cmd.Println("To update:")

	switch method {
	case InstallMethodHomebrew:
		cmd.Println("  brew upgrade kit")
	case InstallMethodGo:
		cmd.Println("  go install github.com/coccinella-labs/go-kit/cmd/kit@latest")
	case InstallMethodScript:
		cmd.Println("  curl -fsSL https://raw.githubusercontent.com/coccinella-labs/go-kit/main/scripts/install.sh | sh")
	case InstallMethodManual:
		cmd.Println("  Download the latest release from:")
		cmd.Println("  https://github.com/coccinella-labs/go-kit/releases/latest")
		cmd.Println()
		cmd.Println("  Then replace the binary manually.")
	default:
		cmd.Println("  Visit https://github.com/coccinella-labs/go-kit/releases/latest")
		cmd.Println("  and download the appropriate binary for your platform.")
	}

	cmd.Println()
	cmd.Println("For more installation options, see: https://github.com/coccinella-labs/go-kit#installation")

	return nil
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "network error") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "Client.Timeout")
}

func getLatestVersion(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 30}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := string(body)
	idx := strings.Index(bodyStr, `"tag_name":"`)
	if idx == -1 {
		return "", fmt.Errorf("tag_name not found in response")
	}

	start := idx + len(`"tag_name":"`)
	end := strings.Index(bodyStr[start:], `"`)
	if end == -1 {
		return "", fmt.Errorf("invalid tag format")
	}

	return bodyStr[start : start+end], nil
}

func detectInstallMethod() InstallMethod {
	if isInHomebrewPath() {
		return InstallMethodHomebrew
	}

	if isInstalledViaGo() {
		return InstallMethodGo
	}

	if isInLocalBin() {
		return InstallMethodScript
	}

	if isInSystemPath() {
		return InstallMethodManual
	}

	return InstallMethodUnknown
}

func isInHomebrewPath() bool {
	brewPrefix, err := runCommand("brew", "--prefix")
	if err != nil {
		return false
	}

	kitBinary, err := os.Executable()
	if err != nil {
		return false
	}

	return strings.HasPrefix(kitBinary, brewPrefix)
}

func isInstalledViaGo() bool {
	kitBinary, err := os.Executable()
	if err != nil {
		return false
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = runtime.GOROOT()
	}

	goBin := goPath + "/bin"
	return strings.HasPrefix(kitBinary, goBin)
}

func isInLocalBin() bool {
	kitBinary, err := os.Executable()
	if err != nil {
		return false
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	localBin := home + "/.local/bin"
	return strings.HasPrefix(kitBinary, localBin)
}

func isInSystemPath() bool {
	kitBinary, err := os.Executable()
	if err != nil {
		return false
	}

	return strings.HasPrefix(kitBinary, "/usr/local/bin") ||
		strings.HasPrefix(kitBinary, "/usr/bin") ||
		strings.HasPrefix(kitBinary, "/opt/homebrew/bin")
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func info(cmd *cobra.Command, msg string) {
	cmd.Printf("[INFO] %s\n", msg)
}

func warn(cmd *cobra.Command, msg string) {
	cmd.Printf("[WARN] %s\n", msg)
}
