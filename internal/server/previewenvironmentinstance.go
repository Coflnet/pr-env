package server

import (
	"fmt"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/labstack/echo/v4"
)

type environmentInstanceModel struct {
	Name            string                         `json:"name"`
	GitSettings     environmentInstanceGitSettings `json:"gitSettings"`
	BuildStatus     string                         `json:"buildStatus"`
	PublicFacingUrl string                         `json:"publicFacingUrl"`
	ContainerImage  string                         `json:"containerImage"`
}

type environmentInstanceGitSettings struct {
	Branch                string `json:"branch"`
	PullRequestIdentifier int    `json:"pullRequestIdentifier"`
	Organization          string `json:"organization"`
	Repository            string `json:"repository"`
	CommitHash            string `json:"commitHash"`
}

func convertToEnvironmentInstanceModelList(in *coflnetv1alpha1.PreviewEnvironmentInstanceList) []environmentInstanceModel {
	res := make([]environmentInstanceModel, len(in.Items))
	for i, v := range in.Items {
		res[i] = convertToEnvironmentInstanceModel(&v)
	}
	return res

}

func convertToEnvironmentInstanceModel(in *coflnetv1alpha1.PreviewEnvironmentInstance) environmentInstanceModel {
	return environmentInstanceModel{
		Name: in.Name,
		GitSettings: environmentInstanceGitSettings{
			Organization:          in.Spec.GitOrganization,
			Repository:            in.Spec.GitRepository,
			Branch:                *in.Spec.Branch,
			CommitHash:            in.Spec.CommitHash,
			PullRequestIdentifier: in.Spec.PullRequestNumber,
		},
		BuildStatus:     in.Status.RebuildStatus,
		PublicFacingUrl: in.Status.PublicFacingUrl,

		// TODO: build the container image from the props
		ContainerImage: fmt.Sprintf("no filled yet"),
	}
}

// List all available Environments
// (GET /environment/list)
func (s Server) GetEnvironmentInstanceListOwner(c echo.Context, owner string) error {
	list, err := s.kubeClient.ListPreviewEnvironmentInstances(c.Request().Context(), owner)
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}

	return c.JSON(200, convertToEnvironmentInstanceModelList(list))
}
