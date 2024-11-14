package git

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v66/github"
)

// ListInstallations lists all installations of the app
// this is just a test function to see if the github app is working
func (g *GithubClient) ListInstallations(ctx context.Context) error {
	installations, _, err := g.appClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		g.log.Info("Installation", "id", installation.GetID(), "mail", installation.GetAccount().GetEmail())
		err := g.ConfigureInstallation(ctx, installation)
		if err != nil {
			return err
		}
	}

	return nil
}

// apiClientForInstallation returns a github client for a specific installation
// if no apiClient for the installation exists a new one is getting created
func (g *GithubClient) apiClientForInstallation(ctx context.Context, installationId int) (*github.Client, error) {
	val, ok := g.userAppClients[installationId]
	token, _ := g.userAppTokens[installationId]

	// check if a applient is already created for the installation
	if ok {

		// if the token is expired, delete the client and token
		// and create a new one
		if token.ExpiresAt.After(time.Now()) {
			delete(g.userAppClients, installationId)
			delete(g.userAppTokens, installationId)
			return g.apiClientForInstallation(ctx, installationId)
		}

		return val, nil
	}

	// create a new client for the installation
	token, _, err := g.appClient.Apps.CreateInstallationToken(ctx, int64(installationId), nil)
	if err != nil {
		return nil, err
	}
	client := github.NewClient(nil).WithAuthToken(token.GetToken())

	// store the client and token
	g.userAppClients[installationId] = client
	g.userAppTokens[installationId] = token

	return client, nil
}

// ListReposOfUser lists all repositories of a user
// uses the github app api to get the repositories
// the installId is the installation id of the user
func (g *GithubClient) ListReposOfUser(ctx context.Context, installId int) (*github.ListRepositories, error) {
	client, err := g.apiClientForInstallation(ctx, installId)
	if err != nil {
		return nil, err
	}

	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (g *GithubClient) ConfigureInstallationById(ctx context.Context, installationId int) error {

	// get the installation
	i, _, err := g.appClient.Apps.GetInstallation(ctx, int64(installationId))
	if err != nil {
		return InstallationIdDoesNotExistError{InstallationId: installationId}
	}

	return g.ConfigureInstallation(ctx, i)
}

func (g *GithubClient) ConfigureInstallation(ctx context.Context, installation *github.Installation) error {
	user, err := g.keycloakClient.UserByGithubId(ctx, int(installation.GetAccount().GetID()))
	if err != nil {
		return err
	}

	// set the installation id as custom attribute in keyclaok
	g.keycloakClient.SetGithubInstallationIdForUser(ctx, *user.ID, int(installation.GetID()))

	return nil
}

type InstallationIdDoesNotExistError struct {
	InstallationId int
}

func (i InstallationIdDoesNotExistError) Error() string {
	return fmt.Sprintf("Installation id %d does not exist", i.InstallationId)
}
