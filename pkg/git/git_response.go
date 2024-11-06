package git

import (
	"context"
	"fmt"
	"time"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/google/go-github/v66/github"
)

func (c *GithubClient) UpdatePullRequestAnswer(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	// NOTE: somehow we don't need this anymore
	// but I am pretty sure we will need the pr in the future
	// first we have to load the pull request
	// pr, err := c.PullRequestOfPei(ctx, pei)
	// if err != nil {
	// 	return err
	// }

	// then we have to figure out what the answer should be
	message := c.messageForPrForPei(pei)

	// message already exists
	doesMessageAlreadyExists := c.pullRequestAlreadyHasMessage(ctx, pe.Spec.GitOrganization, pe.Spec.GitRepository, pei.Spec.PullRequestNumber, message)
	if doesMessageAlreadyExists {
		c.log.Info("Message already exists", "owner", pe.Spec.GitOrganization, "repo", pe.Spec.GitRepository, "prNr", pei.Spec.PullRequestNumber)
		return nil
	}

	// then we have to update the pull request with the answer
	return c.postMessageToPr(ctx, pe.Spec.GitOrganization, pe.Spec.GitRepository, pei.Spec.PullRequestNumber, message)
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
	`, *pei.Spec.Branch, pei.Status.PublicFacingUrl)
}

func (c *GithubClient) postMessageToPr(ctx context.Context, owner, repo string, prNr int, message string) error {
	c.log.Info("Posting message to PR", "owner", owner, "repo", repo, "prNr", prNr, "message", message)
	_, _, err := c.client.Issues.CreateComment(ctx, owner, repo, prNr, &github.IssueComment{
		Body: &message,
	})
	return err
}

func (c *GithubClient) pullRequestAlreadyHasMessage(ctx context.Context, owner, repo string, prNr int, message string) bool {
	c.log.Info("Checking if PR already has message", "owner", owner, "repo", repo, "prNr", prNr, "message", message)
	comments, _, err := c.client.Issues.ListComments(ctx, owner, repo, prNr, &github.IssueListCommentsOptions{})
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
				_, err = c.client.Issues.DeleteComment(ctx, owner, repo, *comment.ID)
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
