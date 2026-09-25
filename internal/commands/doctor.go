package commands

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/coccinella-labs/go-kit/internal/api"
	"github.com/coccinella-labs/go-kit/internal/auth"
	"github.com/coccinella-labs/go-kit/internal/config"
	"github.com/coccinella-labs/go-kit/internal/exitcode"
	"github.com/coccinella-labs/go-kit/internal/git"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose kit setup",
		Long:  `Check configuration, authentication, Git, and network connectivity.`,
		RunE:  runDoctor,
	}

	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
	checks := []struct {
		name string
		run  func() string
	}{
		{"Configuration", checkConfig},
		{"Authentication", checkAuth},
		{"Git", checkGit},
		{"Network", checkNetwork},
	}

	issues := 0
	var advice []string

	for _, c := range checks {
		status := c.run()
		cmd.Printf("%-14s %s\n", c.name, status)
		if status != "OK" {
			issues++
		}
	}

	cmd.Println()
	if issues == 0 {
		cmd.Println("kit is ready.")
		return nil
	}

	for _, c := range checks {
		status := c.run()
		if status != "OK" {
			advice = append(advice, getAdvice(c.name))
		}
	}

	for _, a := range advice {
		cmd.Println(a)
	}

	return exitcode.New(exitcode.ErrGeneral, fmt.Errorf("%d issue(s) found", issues))
}

func checkConfig() string {
	_, err := config.LoadSettings()
	if err != nil {
		return "Failed"
	}
	return "OK"
}

func checkAuth() string {
	if !auth.IsLoggedIn(auth.ProviderGitHub) {
		return "Not signed in"
	}

	token, err := auth.GetToken(auth.ProviderGitHub)
	if err != nil {
		return "Invalid token"
	}

	settings, err := config.LoadSettings()
	if err != nil {
		return "Config error"
	}

	client := api.NewClient(settings.APIURL, token)
	resp, err := client.Get(context.Background(), "/user")
	if err != nil {
		return "Invalid token"
	}
	_ = resp

	return "OK"
}

func checkGit() string {
	_, err := exec.LookPath("git")
	if err != nil {
		return "Not found"
	}

	_, err = git.GetContext()
	if err != nil {
		if err == git.ErrNotInRepo {
			return "OK"
		}
		return "Error"
	}

	return "OK"
}

func checkNetwork() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com", nil)
	if err != nil {
		return "Unreachable"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "Unreachable"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return "OK"
}

func getAdvice(name string) string {
	switch name {
	case "Authentication":
		return "Run:\n  kit auth login github"
	case "Configuration":
		return "Run:\n  kit config list"
	case "Git":
		return "Install Git or run from a repository."
	case "Network":
		return "Check your internet connection."
	default:
		return ""
	}
}
