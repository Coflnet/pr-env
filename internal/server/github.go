package server

import (
	"context"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/labstack/echo/v4"
)

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
