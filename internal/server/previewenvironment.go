package server

import (
	"net/http"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	apigen "github.com/coflnet/pr-env/internal/server/openapi"
	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type environmentModel struct {
	Id                  types.UID                `json:"id"`
	Name                string                   `json:"name"`
	GitSettings         gitSettingsModel         `json:"gitSettings"`
	ContainerSettings   containerSettingsModel   `json:"containerSettings"`
	ApplicationSettings applicationSettingsModel `json:"applicationSettings"`
}

type gitSettingsModel struct {
	Organization string `json:"organization"`
	Repository   string `json:"repository"`
}

type containerSettingsModel struct {
	Registry   string `json:"registry"`
	Repository string `json:"image"`
}

type applicationSettingsModel struct {
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
}

// List all available Environments
// (GET /environment/list)
func (s Server) GetEnvironmentList(c echo.Context, p apigen.GetEnvironmentListParams) error {
	list, err := s.kubeClient.ListPreviewEnvironments(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, convertToEnvironmentModelList(list))
}

// Creates a new environment
// (POST /environment)
func (s Server) PostEnvironment(c echo.Context, p apigen.PostEnvironmentParams) error {
	var env environmentModel
	if err := c.Bind(&env); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// TODO: check if the user already has an environment with the same name
	// kPe, err := s.kubeClient.PreviewEnvironmentById(c.Request().Context(), env.Name)
	// if err != nil {
	// 	if client.IgnoreNotFound(err) != nil {
	// 		s.log.Error(err, "failed to get environment by name")
	// 		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// 	}
	// }
	// if kPe != nil {
	// 	s.log.Info("environment already exists", "name", env.Name)
	// 	return echo.NewHTTPError(http.StatusConflict, "environment already exists")
	// }

	// convert the model to the CRD
	pe, err := convertFromEnvironment(&env)
	if err != nil {
		s.log.Error(err, "failed to convert environment model")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// create the environment
	if err := s.kubeClient.CreatePreviewEnvironment(c.Request().Context(), pe); err != nil {
		s.log.Error(err, "failed to create environment")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// load the created environment
	// TODO: fix this
	pe, err = s.kubeClient.PreviewEnvironmentById(c.Request().Context(), types.UID(""), "")
	if err != nil {
		s.log.Error(err, "failed to get environment by name after creation")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, convertToEnvironmentModel(pe))
}

// Deletes an environment
// (DELETE /environment/{name})
func (s Server) DeleteEnvironmentOwnerId(c echo.Context, owner string, id string, p apigen.DeleteEnvironmentOwnerIdParams) error {
	// delete the environment
	pe, err := s.kubeClient.DeletePreviewEnvironment(c.Request().Context(), types.UID(id), owner)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			s.log.Error(err, "failed to delete environment")
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		s.log.Info("environment not found", "id", id, "owner", owner)
		return echo.NewHTTPError(http.StatusNotFound, "environment not found")
	}

	return c.JSON(http.StatusOK, convertToEnvironmentModel(pe))
}

func convertToEnvironmentModel(in *coflnetv1alpha1.PreviewEnvironment) *environmentModel {
	return &environmentModel{
		Id:   in.GetUID(),
		Name: in.Name,
		GitSettings: gitSettingsModel{
			Organization: in.Spec.GitOrganization,
			Repository:   in.Spec.GitRepository,
		},
		ContainerSettings: containerSettingsModel{
			Registry:   in.Spec.ContainerRegistry.Registry,
			Repository: in.Spec.ContainerRegistry.Repository,
		},
		ApplicationSettings: applicationSettingsModel{
			Port:     in.Spec.ApplicationSettings.Port,
			Hostname: in.Spec.ApplicationSettings.IngressHostname,
		},
	}
}

func convertToEnvironmentModelList(in *coflnetv1alpha1.PreviewEnvironmentList) []*environmentModel {
	out := make([]*environmentModel, len(in.Items))
	for i, item := range in.Items {
		out[i] = convertToEnvironmentModel(&item)
	}
	return out
}

func convertFromEnvironment(in *environmentModel) (*coflnetv1alpha1.PreviewEnvironment, error) {
	return &coflnetv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name: in.Name,
		},
		Spec: coflnetv1alpha1.PreviewEnvironmentSpec{
			GitOrganization: in.GitSettings.Organization,
			GitRepository:   in.GitSettings.Repository,
			ContainerRegistry: coflnetv1alpha1.ContainerRegistry{
				Registry:   in.ContainerSettings.Registry,
				Repository: in.ContainerSettings.Repository,
			},
			ApplicationSettings: coflnetv1alpha1.ApplicationSettings{
				Port:            in.ApplicationSettings.Port,
				IngressHostname: in.ApplicationSettings.Hostname,
			},
		},
	}, nil
}
