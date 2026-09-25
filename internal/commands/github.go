package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/coccinella-labs/go-kit/internal/api"
	"github.com/coccinella-labs/go-kit/internal/auth"
	"github.com/coccinella-labs/go-kit/internal/config"
	"github.com/coccinella-labs/go-kit/internal/git"
	"github.com/spf13/cobra"
)

func newGithubCmd() *cobra.Command {
	githubCmd := &cobra.Command{
		Use:   "github",
		Short: "GitHub integration",
		Long:  `Interact with GitHub repositories, issues, and pull requests.`,
	}

	githubCmd.AddCommand(newGithubRepoCmd())
	githubCmd.AddCommand(newGithubIssueCmd())
	githubCmd.AddCommand(newGithubPrCmd())

	return githubCmd
}

func newGithubRepoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository commands",
		Long:  `Manage and list GitHub repositories.`,
	}

	repoCmd.AddCommand(newGithubRepoListCmd())
	repoCmd.AddCommand(newGithubRepoCloneCmd())

	return repoCmd
}

func newGithubRepoCloneCmd() *cobra.Command {
	var owner, repo string

	cmd := &cobra.Command{
		Use:   "clone <owner/repo>",
		Short: "Clone a repository",
		Long:  `Clone a GitHub repository to the current directory.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid repository format: expected owner/repo")
			}
			owner = parts[0]
			repo = parts[1]

			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewClient(settings.APIURL, token)

			resp, err := client.Get(cmd.Context(), fmt.Sprintf("/repos/%s/%s", owner, repo))
			if err != nil {
				return fmt.Errorf("failed to get repository: %w", err)
			}

			var repoData map[string]interface{}
			if err := resp.Decode(&repoData); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			cloneURL, _ := repoData["clone_url"].(string)
			if cloneURL == "" {
				return fmt.Errorf("repository not found or no clone URL")
			}

			cmd.Printf("Cloning %s/%s...\n", owner, repo)

			cloneCmd := exec.Command("git", "clone", cloneURL)
			cloneCmd.Stdout = cmd.OutOrStdout()
			cloneCmd.Stderr = cmd.OutOrStderr()
			cloneCmd.Dir = "."

			if err := cloneCmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}

			cmd.Printf("Successfully cloned %s/%s\n", owner, repo)
			return nil
		},
	}

	return cmd
}

func newGithubRepoListCmd() *cobra.Command {
	var perPage int
	var page int
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		Long:  `List repositories accessible to the authenticated user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewClient(settings.APIURL, token)

			params := map[string]string{
				"per_page": fmt.Sprintf("%d", perPage),
				"page":     fmt.Sprintf("%d", page),
			}

			resp, err := client.Get(cmd.Context(), "/user/repos", api.WithQueryParams(params))
			if err != nil {
				return fmt.Errorf("failed to list repos: %w", err)
			}

			if jsonOutput {
				cmd.Println(string(resp.Body))
				return nil
			}

			var repos []map[string]interface{}
			if err := resp.Decode(&repos); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			for _, repo := range repos {
				if quiet {
					name, _ := repo["full_name"].(string)
					cmd.Println(name)
				} else {
					fullName, _ := repo["full_name"].(string)
					visibility, _ := repo["visibility"].(string)
					updatedAt, _ := repo["updated_at"].(string)
					cmd.Printf("%s (%s) updated %s\n", fullName, visibility, updatedAt)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&perPage, "per-page", "n", 30, "number of results per page")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "page number")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output raw JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "minimal output")

	return cmd
}

func newGithubIssueCmd() *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue commands",
		Long:  `Manage and list GitHub issues.`,
	}

	issueCmd.AddCommand(newGithubIssueListCmd())
	issueCmd.AddCommand(newGithubIssueCreateCmd())

	return issueCmd
}

func newGithubIssueCreateCmd() *cobra.Command {
	var owner, repo, title, body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue",
		Long:  `Create a new GitHub issue.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if owner == "" || repo == "" {
				ctx, gitErr := git.GetContext()
				if gitErr == nil {
					if owner == "" {
						owner = ctx.Owner
					}
					if repo == "" {
						repo = ctx.Repo
					}
				}
			}

			if owner == "" || repo == "" || title == "" {
				return fmt.Errorf("owner, repo, and title are required")
			}

			client := api.NewClient(settings.APIURL, token)

			issueBody := map[string]interface{}{
				"title": title,
				"body":  body,
			}

			resp, err := client.Post(cmd.Context(), fmt.Sprintf("/repos/%s/%s/issues", owner, repo), issueBody)
			if err != nil {
				return fmt.Errorf("failed to create issue: %w", err)
			}

			var issue map[string]interface{}
			if err := resp.Decode(&issue); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			number, _ := issue["number"].(float64)
			issueURL, _ := issue["html_url"].(string)

			cmd.Printf("Created issue #%d: %s\n", int(number), issueURL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&owner, "owner", "o", "", "repository owner")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository name")
	cmd.Flags().StringVarP(&title, "title", "t", "", "issue title")
	cmd.Flags().StringVarP(&body, "body", "b", "", "issue body")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}

func newGithubIssueListCmd() *cobra.Command {
	var perPage int
	var page int
	var state string
	var owner, repo string
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		Long:  `List issues for a repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if owner == "" || repo == "" {
				ctx, gitErr := git.GetContext()
				if gitErr == nil {
					if owner == "" {
						owner = ctx.Owner
					}
					if repo == "" {
						repo = ctx.Repo
					}
				}
			}

			if owner == "" || repo == "" {
				return fmt.Errorf("owner and repo are required (use --owner and --repo flags, or run from a git repository)")
			}

			client := api.NewClient(settings.APIURL, token)

			params := map[string]string{
				"per_page": fmt.Sprintf("%d", perPage),
				"page":     fmt.Sprintf("%d", page),
				"state":    state,
			}

			path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
			resp, err := client.Get(cmd.Context(), path, api.WithQueryParams(params))
			if err != nil {
				return fmt.Errorf("failed to list issues: %w", err)
			}

			if jsonOutput {
				cmd.Println(string(resp.Body))
				return nil
			}

			var issues []map[string]interface{}
			if err := resp.Decode(&issues); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			for _, issue := range issues {
				if quiet {
					number, _ := issue["number"].(float64)
					cmd.Printf("%d\n", int(number))
				} else {
					number, _ := issue["number"].(float64)
					title, _ := issue["title"].(string)
					state, _ := issue["state"].(string)
					cmd.Printf("#%d [%s] %s\n", int(number), state, title)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&perPage, "per-page", "n", 30, "number of results per page")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "page number")
	cmd.Flags().StringVarP(&state, "state", "s", "open", "issue state: open, closed, all")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "repository owner")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository name")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output raw JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "minimal output")

	return cmd
}

func newGithubPrCmd() *cobra.Command {
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Pull request commands",
		Long:  `Manage and list GitHub pull requests.`,
	}

	prCmd.AddCommand(newGithubPrListCmd())
	prCmd.AddCommand(newGithubPrCreateCmd())
	prCmd.AddCommand(newGithubPrCheckoutCmd())

	return prCmd
}

func newGithubPrCreateCmd() *cobra.Command {
	var owner, repo, title, body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		Long:  `Create a new pull request.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if owner == "" || repo == "" {
				ctx, gitErr := git.GetContext()
				if gitErr == nil {
					if owner == "" {
						owner = ctx.Owner
					}
					if repo == "" {
						repo = ctx.Repo
					}
				}
			}

			if owner == "" || repo == "" || title == "" {
				return fmt.Errorf("owner, repo, and title are required")
			}

			client := api.NewClient(settings.APIURL, token)

			prBody := map[string]interface{}{
				"title": title,
				"body":  body,
			}

			resp, err := client.Post(cmd.Context(), fmt.Sprintf("/repos/%s/%s/pulls", owner, repo), prBody)
			if err != nil {
				return fmt.Errorf("failed to create pull request: %w", err)
			}

			var pr map[string]interface{}
			if err := resp.Decode(&pr); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			number, _ := pr["number"].(float64)
			prURL, _ := pr["html_url"].(string)

			cmd.Printf("Created PR #%d: %s\n", int(number), prURL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&owner, "owner", "o", "", "repository owner")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository name")
	cmd.Flags().StringVarP(&title, "title", "t", "", "pull request title")
	cmd.Flags().StringVarP(&body, "body", "b", "", "pull request body")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}

func newGithubPrCheckoutCmd() *cobra.Command {
	var owner, repo string
	var number int

	cmd := &cobra.Command{
		Use:   "checkout <number>",
		Short: "Checkout a pull request branch",
		Long:  `Checkout the branch of a pull request locally.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid PR number: %s", args[0])
			}
			number = num

			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if owner == "" || repo == "" {
				ctx, gitErr := git.GetContext()
				if gitErr == nil {
					if owner == "" {
						owner = ctx.Owner
					}
					if repo == "" {
						repo = ctx.Repo
					}
				}
			}

			if owner == "" || repo == "" {
				return fmt.Errorf("owner and repo are required (use --owner and --repo flags, or run from a git repository)")
			}

			client := api.NewClient(settings.APIURL, token)

			resp, err := client.Get(cmd.Context(), fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number))
			if err != nil {
				return fmt.Errorf("failed to get pull request: %w", err)
			}

			var pr map[string]interface{}
			if err := resp.Decode(&pr); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			head, _ := pr["head"].(map[string]interface{})
			if head == nil {
				return fmt.Errorf("invalid pull request data")
			}

			headRepo, _ := head["repo"].(map[string]interface{})
			headRef, _ := head["ref"].(string)
			headLogin, _ := head["user"].(map[string]interface{})["login"].(string)

			if headRepo == nil {
				return fmt.Errorf("pull request head repository not found")
			}

			cloneURL, _ := headRepo["clone_url"].(string)
			if cloneURL == "" {
				return fmt.Errorf("no clone URL for pull request head")
			}

			branchName := fmt.Sprintf("pr-%d-%s", number, headRef)
			if headLogin != "" {
				branchName = fmt.Sprintf("%s-pr-%d-%s", headLogin, number, headRef)
			}

			cmd.Printf("Checking out PR #%d into branch '%s'...\n", number, branchName)

			fetchCmd := exec.Command("git", "fetch", "origin", fmt.Sprintf("pull/%d/head:%s", number, branchName))
			fetchCmd.Dir = "."
			fetchCmd.Stdout = cmd.OutOrStdout()
			fetchCmd.Stderr = cmd.OutOrStderr()

			if err := fetchCmd.Run(); err != nil {
				return fmt.Errorf("git fetch failed: %w", err)
			}

			checkoutCmd := exec.Command("git", "checkout", branchName)
			checkoutCmd.Dir = "."
			checkoutCmd.Stdout = cmd.OutOrStdout()
			checkoutCmd.Stderr = cmd.OutOrStderr()

			if err := checkoutCmd.Run(); err != nil {
				return fmt.Errorf("git checkout failed: %w", err)
			}

			cmd.Printf("Checked out PR #%d into branch '%s'\n", number, branchName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&owner, "owner", "o", "", "repository owner")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository name")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

func newGithubPrListCmd() *cobra.Command {
	var perPage int
	var page int
	var state string
	var owner, repo string
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		Long:  `List pull requests for a repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.GetToken(auth.ProviderGitHub)
			if err != nil {
				return fmt.Errorf("not authenticated: run 'kit auth login github --token <token>'")
			}

			settings, err := config.LoadSettings()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if owner == "" || repo == "" {
				ctx, gitErr := git.GetContext()
				if gitErr == nil {
					if owner == "" {
						owner = ctx.Owner
					}
					if repo == "" {
						repo = ctx.Repo
					}
				}
			}

			if owner == "" || repo == "" {
				return fmt.Errorf("owner and repo are required (use --owner and --repo flags, or run from a git repository)")
			}

			client := api.NewClient(settings.APIURL, token)

			params := map[string]string{
				"per_page": fmt.Sprintf("%d", perPage),
				"page":     fmt.Sprintf("%d", page),
				"state":    state,
			}

			path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
			resp, err := client.Get(cmd.Context(), path, api.WithQueryParams(params))
			if err != nil {
				return fmt.Errorf("failed to list pull requests: %w", err)
			}

			if jsonOutput {
				cmd.Println(string(resp.Body))
				return nil
			}

			var prs []map[string]interface{}
			if err := resp.Decode(&prs); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			for _, pr := range prs {
				if quiet {
					number, _ := pr["number"].(float64)
					cmd.Printf("%d\n", int(number))
				} else {
					number, _ := pr["number"].(float64)
					title, _ := pr["title"].(string)
					state, _ := pr["state"].(string)
					cmd.Printf("#%d [%s] %s\n", int(number), state, title)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&perPage, "per-page", "n", 30, "number of results per page")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "page number")
	cmd.Flags().StringVarP(&state, "state", "s", "open", "PR state: open, closed, all")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "repository owner")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository name")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output raw JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "minimal output")

	return cmd
}
