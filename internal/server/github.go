package server

import (
	"context"
	"strconv"
	"strings"

	"github.com/coflnet/pr-env/internal/git"
	"github.com/coflnet/pr-env/internal/keycloak"
	apigen "github.com/coflnet/pr-env/internal/server/openapi"
	"github.com/google/go-github/v66/github"
	"github.com/labstack/echo/v4"
)

func convertToGithubRepositoryModelList(repos []*github.Repository) []apigen.ServerGithubRepositoryModel {
	res := make([]apigen.ServerGithubRepositoryModel, len(repos))
	for i, repo := range repos {
		res[i] = convertToGithubRepositoryModel(repo)
	}
	return res
}

func convertToGithubRepositoryModel(repo *github.Repository) apigen.ServerGithubRepositoryModel {
	return apigen.ServerGithubRepositoryModel{
		Name:  strPtr(repo.GetName()),
		Owner: strPtr(repo.GetOwner().GetLogin()),
	}
}

func strPtr(s string) *string {
	return &s
}

func (s Server) GetGithubRepositories(c echo.Context, params apigen.GetGithubRepositoriesParams) error {
	owner := c.Get("user").(string)
	s.log.Info("Listing Github repositories", "owner", owner)

	installationId, err := s.keycloakClient.GithubInstallationIdForUser(c.Request().Context(), owner)
	if err != nil {
		if e, ok := err.(keycloak.InstallationIdDoesNotExistError); ok {
			s.log.Info("Installation id does not exist", "user", e.UserId)
			return echo.NewHTTPError(401, "User has no github app connected")
		}

		s.log.Error(err, "Unable to get Github installation id")
		return echo.NewHTTPError(500, err.Error())
	}
	s.log.Info("Found Github installation id", "id", installationId)

	repos, err := s.githubClient.ListReposOfUser(c.Request().Context(), installationId)
	if err != nil {
		s.log.Error(err, "Unable to list repositories")
		return echo.NewHTTPError(500, err.Error())
	}

	s.log.Info("Found repositories", "count", len(repos.Repositories))
	return c.JSON(200, convertToGithubRepositoryModelList(repos.Repositories))
}

func (s Server) ConfigureInstallation(c echo.Context) error {
	installationIdStr := c.QueryParam("installation_id")
	if installationIdStr == "" {
		return echo.NewHTTPError(400, "installation_id missing")
	}

	installationId, err := strconv.Atoi(installationIdStr)
	if err != nil {
		return echo.NewHTTPError(400, "installation_id is not an integer")
	}

	s.log.Info("Configuring installation", "id", installationId)
	err = s.githubClient.ConfigureInstallationById(c.Request().Context(), installationId)
	if err != nil {
		if e, ok := err.(git.InstallationIdDoesNotExistError); ok {
			s.log.Error(e, "Installation id does not exist")
			return echo.NewHTTPError(404, e.Error())
		}

		s.log.Error(err, "Unable to configure installation")
		return echo.NewHTTPError(500, err.Error())
	}

	s.log.Info("Installation configured", "id", installationId)
	return c.JSON(200, "ok")
}

func (s *Server) GithubWebhookHandler(c echo.Context) error {
	s.log.Info("Received Github Webhook")

	payload, err := github.ValidatePayload(c.Request(), []byte(""))
	if err != nil {
		s.log.Error(err, "Unable to validate payload")
		return err
	}

	t := github.WebHookType(c.Request())
	event, err := github.ParseWebHook(t, payload)
	if err != nil {
		s.log.Error(err, "Unable to parse")
		return err
	}

	switch event := event.(type) {
	case *github.PushEvent:
		err := s.HandleGithubPush(c.Request().Context(), event)
		if err != nil {
			s.log.Error(err, "Unable to handle push event")
		}
		return err
	case *github.PullRequestEvent:
		err := s.HandleGithubPullRequest(c.Request().Context(), event)
		if err != nil {
			s.log.Error(err, "Unable to handle pull request event")
		}
		return err
	default:
		s.log.Info("Received event but not an important one", "type", t)
		return nil
	}
}

func (s *Server) HandleGithubPush(ctx context.Context, event *github.PushEvent) error {
	ref := *event.Ref
	branch := strings.Split(ref, "/")[2]

	owner, repo := *event.Repo.Owner.Name, *event.Repo.Name
	prs, err := s.githubClient.PullRequestsOfRepositoryAndBranch(ctx, owner, repo, branch)
	if err != nil {
		return err
	}

	s.log.Info("Searched pull requests for branch", "owner", owner, "repo", repo, "branch", branch, "prs", len(prs))

	for _, pr := range prs {
		err = s.HandleGithubEvent(ctx, owner, repo, *pr.Number)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) HandleGithubPullRequest(ctx context.Context, event *github.PullRequestEvent) error {
	return s.HandleGithubEvent(ctx, *event.Repo.Owner.Name, *event.Repo.Name, *event.PullRequest.Number)
}

func (s *Server) HandleGithubEvent(ctx context.Context, owner, repo string, prNumber int) error {
	s.log.Info("Handling Github event", "owner", owner, "repo", repo)
	return s.kubeClient.TriggerUpdateForPreviewEnvironmentInstance(ctx, owner, repo, prNumber)
}
