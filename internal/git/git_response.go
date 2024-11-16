package git

import (
	"context"
	"fmt"
	"time"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/google/go-github/v66/github"
)

func (c *GithubClient) UpdatePullRequestAnswer(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	// then we have to figure out what the answer should be
	message := c.messageForPrForPei(pei)

	// message already exists
	doesMessageAlreadyExists := c.pullRequestAlreadyHasMessage(ctx, pe.Spec.GitSettings.Organization, pe.Spec.GitSettings.Repository, *pei.Spec.InstanceGitSettings.PullRequestNumber, message)
	if doesMessageAlreadyExists {
		c.log.Info("Message already exists", "owner", pe.Spec.GitSettings.Organization, "repo", pe.Spec.GitSettings.Repository, "prNr", pei.Spec.InstanceGitSettings.PullRequestNumber)
		return nil
	}

	// then we have to update the pull request with the answer
	return c.postMessageToPr(ctx, pe.Spec.GitSettings.Organization, pe.Spec.GitSettings.Repository, *pei.Spec.InstanceGitSettings.PullRequestNumber, message)
}

func (c *GithubClient) messageForPrForPei(pei *coflnetv1alpha1.PreviewEnvironmentInstance) string {
	return fmt.Sprintf(`
Hello! This is an automated message from the Preview Environment Operator.
We have detected that there should be a new preview environment for the branch %s.

---

Here you can checkout a preview version for the changes:
%s

---

Maybe you can already find some issues here.
If not even better!
	`, *pei.Spec.InstanceGitSettings.Branch, pei.Status.PublicFacingUrl)
}

func (c *GithubClient) postMessageToPr(ctx context.Context, owner, repo string, prNr int, message string) error {
	c.log.Info("Posting message to PR", "owner", owner, "repo", repo, "prNr", prNr, "message", message)
	_, _, err := c.oauthClient.Issues.CreateComment(ctx, owner, repo, prNr, &github.IssueComment{
		Body: &message,
	})
	return err
}

func (c *GithubClient) pullRequestAlreadyHasMessage(ctx context.Context, owner, repo string, prNr int, message string) bool {
	c.log.Info("Checking if PR already has message", "owner", owner, "repo", repo, "prNr", prNr, "message", message)
	comments, _, err := c.oauthClient.Issues.ListComments(ctx, owner, repo, prNr, &github.IssueListCommentsOptions{})
	if err != nil {
		c.log.Error(err, "unable to list comments")
		return false
	}

	for _, comment := range comments {
		if *comment.Body == message {

			// check if the message should be refershed
			// TODO: it would be better to check for the last commit hash and update based on that
			// but for the development version this is enough
			age := time.Now().Sub(comment.CreatedAt.Time)
			if age > time.Hour*24 {
				_, err = c.oauthClient.Issues.DeleteComment(ctx, owner, repo, *comment.ID)
				if err != nil {
					c.log.Error(err, "unable to delete comment, but it is outdated")
					return false
				}
				return false
			}
			return true
		}
	}
	return false
}
