package git

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	coflnetv1alpha "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/coflnet/pr-env/internal/keycloak"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v66/github"
	_ "github.com/joho/godotenv/autoload"
)

var (
	githubOauthClientInstance *github.Client
)

func NewGithubClient(logger logr.Logger, keycloak *keycloak.KeycloakClient) (*GithubClient, error) {
	oauthClient := githubOauthClient()
	appClient, err := githubAppClient()
	if err != nil {
		return nil, err
	}

	return &GithubClient{
		log:            logger,
		oauthClient:    oauthClient,
		appClient:      appClient,
		keycloakClient: keycloak,

		userAppClients: make(map[int]*github.Client),
		userAppTokens:  make(map[int]*github.InstallationToken),
	}, nil
}

func githubOauthClient() *github.Client {
	if githubOauthClientInstance != nil {
		return githubOauthClientInstance
	}

	if authTokenSet() {
		return github.NewClient(nil).WithAuthToken(authToken())
	}

	return github.NewClient(nil)
}

func githubAppClient() (*github.Client, error) {
	tr := http.DefaultTransport

	privatePem, err := os.ReadFile(githubAppPrivateKeyPath())
	if err != nil {
		return nil, err
	}

	itr, err := ghinstallation.NewAppsTransport(tr, 1054539, privatePem)
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}

type GithubClient struct {
	oauthClient    *github.Client
	appClient      *github.Client
	keycloakClient *keycloak.KeycloakClient
	log            logr.Logger

	userAppClients map[int]*github.Client
	userAppTokens  map[int]*github.InstallationToken
}

func (c *GithubClient) PullRequestsOfRepository(ctx context.Context, owner, repo string) ([]*github.PullRequest, error) {
	prs, _, err := c.oauthClient.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (c *GithubClient) PullRequestsOfRepositoryAndBranch(ctx context.Context, owner, repo, branch string) ([]*github.PullRequest, error) {
	prs, _, err := c.oauthClient.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{})
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

func (c *GithubClient) PullRequestOfPei(ctx context.Context, pei *coflnetv1alpha.PreviewEnvironmentInstance) (*github.PullRequest, error) {
	return c.PullRequest(ctx, pei.Spec.GitOrganization, pei.Spec.GitRepository, pei.Spec.PullRequestNumber)
}

func (c *GithubClient) PullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.oauthClient.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (c *GithubClient) BranchesOfRepository(ctx context.Context, owner, repo string) ([]string, error) {
	branches, _, err := c.oauthClient.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{})
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

func githubAppPrivateKeyPath() string {
	v := os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH")
	if v == "" {
		panic("GITHUB_APP_PRIVATE_KEY_PATH is not set")
	}
	return v
}
