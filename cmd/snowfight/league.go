package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// runLeague fetches up to 1000 repositories matching sfc-snowbot-*, finds the first .js file in each, and prints its raw URL.
func runLeague(args []string) error {
	// Token for authentication (optional but helps with rate limits)
	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	// Search for repositories matching sfc-snowbot-*
	var repos []*github.Repository
	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}}
	query := "sfc-snowbot- in:name"

	for {
		result, resp, err := client.Search.Repositories(ctx, query, opt)
		if err != nil {
			return fmt.Errorf("searching repositories: %w", err)
		}
		repos = append(repos, result.Repositories...)
		if len(repos) >= 1000 || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Limit to 1000
	if len(repos) > 1000 {
		repos = repos[:1000]
	}

	// Process each repository
	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()
		branch := repo.GetDefaultBranch()
		// List root contents
		_, dirContents, _, err := client.Repositories.GetContents(ctx, owner, name, "", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot list contents of %s: %v\n", name, err)
			continue
		}
		var jsFiles []string
		for _, entry := range dirContents {
			if entry.GetType() == "file" && strings.HasSuffix(entry.GetName(), ".js") {
				jsFiles = append(jsFiles, entry.GetPath())
			}
		}
		if len(jsFiles) == 0 {
			continue
		}
		sort.Strings(jsFiles)
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, name, branch, jsFiles[0])
		fmt.Println(rawURL)
	}
	return nil
}
