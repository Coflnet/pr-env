package git

import (
	"context"
	"os"
	"strings"

	"github.com/google/go-github/v66/github"
	_ "github.com/joho/godotenv/autoload"
)

var githubClient *GithubClient

func NewGithubClient() (*GithubClient, error) {
	if githubClient != nil {
		return githubClient, nil
	}

	if authTokenSet() {
		githubClient = &GithubClient{
			client: github.NewClient(nil).WithAuthToken(authToken()),
		}
	} else {
		githubClient = &GithubClient{
			client: github.NewClient(nil),
		}
	}

	return githubClient, nil
}

type GithubClient struct {
	client *github.Client
}

func (c *GithubClient) PullRequestsOfRepository(ctx context.Context, owner, repo string) ([]*github.PullRequest, error) {
	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (c *GithubClient) PullRequestsOfRepositoryAndBranch(ctx context.Context, owner, repo, branch string) ([]*github.PullRequest, error) {
	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredPrs []*github.PullRequest
	for _, pr := range prs {
		if strings.Contains(pr.GetHead().GetRef(), branch) {
			filteredPrs = append(filteredPrs, pr)
		}
	}

	return filteredPrs, nil
}

func (c *GithubClient) PullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (c *GithubClient) BranchesOfRepository(ctx context.Context, owner, repo string) ([]string, error) {
	branches, _, err := c.client.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{})
	if err != nil {
		return nil, err
	}

	var branchNames []string
	for _, branch := range branches {
		b := branch.GetName()
		if b == "" {
			continue
		}

		branchNames = append(branchNames, b)
	}

	return branchNames, nil
}

func authToken() string {
	return os.Getenv("GITHUB_AUTH_TOKEN")
}

func authTokenSet() bool {
	return authToken() != ""
}
