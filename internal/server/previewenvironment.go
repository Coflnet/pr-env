package server

import (
	"context"
	"fmt"
	"net/http"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	apigen "github.com/coflnet/pr-env/internal/server/openapi"
	"github.com/labstack/echo/v4"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// List all available Environments
// (GET /environment/list)
func (s Server) GetEnvironmentList(ctx context.Context, request apigen.GetEnvironmentListRequestObject) (apigen.GetEnvironmentListResponseObject, error) {
	owner, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	list, err := s.kubeClient.ListPreviewEnvironments(ctx, owner)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return convertToEnvironmentModelList(list), nil
}

// Creates a new environment
// (POST /environment)
func (s Server) PostEnvironment(ctx context.Context, request apigen.PostEnvironmentRequestObject) (apigen.PostEnvironmentResponseObject, error) {
	userId, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	err = s.kubeClient.CreatePreviewEnvironment(ctx, convertFromEnvironmentModel(userId, *request.Body))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	name := coflnetv1alpha1.PreviewEnvironmentName(request.Body.GitSettings.Organization, request.Body.GitSettings.Repository)
	pe, err := s.kubeClient.PreviewEnvironmentByName(ctx, userId, name)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return apigen.PostEnvironment200JSONResponse(convertToEnvironmentModel(pe)), nil
}

// Deletes an environment
// (DELETE /environment/{id})
func (s Server) DeleteEnvironmentId(ctx context.Context, request apigen.DeleteEnvironmentIdRequestObject) (apigen.DeleteEnvironmentIdResponseObject, error) {
	userId, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	s.log.Info("Deleting PreviewEnvironment", "id", request.Id)
	pe, err := s.kubeClient.DeletePreviewEnvironment(ctx, userId, types.UID(request.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	s.log.Info("Deleted PreviewEnvironment", "name", pe.GetName())
	return apigen.DeleteEnvironmentId200JSONResponse(convertToEnvironmentModel(pe)), nil

}

// Add a user to an environment
// (PATCH /environment/addUser/{id})
func (s Server) PatchEnvironmentAddUserEnvironmentIdUserId(ctx context.Context, request apigen.PatchEnvironmentAddUserEnvironmentIdUserIdRequestObject) (apigen.PatchEnvironmentAddUserEnvironmentIdUserIdResponseObject, error) {
	userId, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	pe, err := s.kubeClient.PreviewEnvironmentById(ctx, userId, types.UID(request.EnvironmentId))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("environment with id %s not found", request.EnvironmentId))
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	for _, u := range pe.Spec.AccessSettings.Users {
		if u.UserId == request.UserId {
			return nil, echo.NewHTTPError(http.StatusConflict, fmt.Errorf("user with id %s already exists", request.UserId))
		}
	}

	pe.Spec.AccessSettings.Users = append(pe.Spec.AccessSettings.Users, coflnetv1alpha1.UserAccess{
		UserId: request.UserId,
	})

	err = s.kubeClient.UpdatePreviewEnvironment(ctx, pe)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return apigen.PatchEnvironmentAddUserEnvironmentIdUserId200JSONResponse(convertToEnvironmentModel(pe)), nil
}

// Remove a user from an environment
// (PATCH /environment/removeUser/{id})
func (s Server) PatchEnvironmentRemoveUserEnvironmentIdUserId(ctx context.Context, request apigen.PatchEnvironmentRemoveUserEnvironmentIdUserIdRequestObject) (apigen.PatchEnvironmentRemoveUserEnvironmentIdUserIdResponseObject, error) {
	userId, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	pe, err := s.kubeClient.PreviewEnvironmentById(ctx, userId, types.UID(request.EnvironmentId))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("environment with id %s not found", request.EnvironmentId))
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	for i, u := range pe.Spec.AccessSettings.Users {
		if u.UserId == request.UserId {
			pe.Spec.AccessSettings.Users = append(pe.Spec.AccessSettings.Users[:i], pe.Spec.AccessSettings.Users[i+1:]...)
			break
		}
	}

	err = s.kubeClient.UpdatePreviewEnvironment(ctx, pe)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return apigen.PatchEnvironmentRemoveUserEnvironmentIdUserId200JSONResponse(convertToEnvironmentModel(pe)), nil
}

func convertFromEnvironmentModel(userId string, in apigen.PreviewEnvironmentModel) *coflnetv1alpha1.PreviewEnvironment {
	name := coflnetv1alpha1.PreviewEnvironmentName(in.GitSettings.Organization, in.GitSettings.Repository)
	vars := make([]coflnetv1alpha1.EnvironmentVariable, 0)

	if in.ApplicationSettings.EnvironmentVariables != nil {
		for _, v := range *in.ApplicationSettings.EnvironmentVariables {
			vars = append(vars, coflnetv1alpha1.EnvironmentVariable{
				Key:   v.Key,
				Value: v.Value,
			})
		}
	}

	users := []coflnetv1alpha1.UserAccess{}
	for _, u := range in.AccessSettings.Users {
		users = append(users, coflnetv1alpha1.UserAccess{
			UserId:   u.UserId,
			Username: u.Username,
		})
	}

	return &coflnetv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"owner": userId,
			},
		},
		Spec: coflnetv1alpha1.PreviewEnvironmentSpec{
			AccessSettings: coflnetv1alpha1.AccessSettings{
				Users:        users,
				PublicAccess: false,
			},
			ApplicationSettings: coflnetv1alpha1.ApplicationSettings{
				Command:              in.ApplicationSettings.Command,
				EnvironmentVariables: &vars,
				Port:                 in.ApplicationSettings.Port,
				IngressHostname:      "tmpenv.app",
			},
			BuildSettings: coflnetv1alpha1.BuildSettings{
				BranchWildcard:       in.BuildSettings.BranchWildcard,
				BuildAllBranches:     in.BuildSettings.BuildAllBranches,
				BuildAllPullRequests: in.BuildSettings.BuildAllPullRequests,
				DockerfilePath:       in.BuildSettings.DockerFilePath,
			},
			ContainerRegistry: &coflnetv1alpha1.ContainerRegistry{
				Registry:   "index.docker.io",
				Repository: "muehlhansfl",
			},
			GitSettings: coflnetv1alpha1.GitSettings{
				Organization: in.GitSettings.Organization,
				Repository:   in.GitSettings.Repository,
			},
		},
	}
}

func convertToEnvironmentModelList(in *coflnetv1alpha1.PreviewEnvironmentList) apigen.GetEnvironmentList200JSONResponse {
	out := make([]apigen.PreviewEnvironmentModel, len(in.Items))
	for i, item := range in.Items {
		out[i] = convertToEnvironmentModel(&item)
	}
	return out
}

func convertToEnvironmentModel(in *coflnetv1alpha1.PreviewEnvironment) apigen.PreviewEnvironmentModel {
	vars := make([]apigen.EnvironmentVariableModel, len(*in.Spec.ApplicationSettings.EnvironmentVariables))
	for i, v := range *in.Spec.ApplicationSettings.EnvironmentVariables {
		vars[i] = apigen.EnvironmentVariableModel{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	users := make([]struct {
		UserId   string `json:"userId"`
		Username string `json:"username"`
	}, len(in.Spec.AccessSettings.Users))

	for i, u := range in.Spec.AccessSettings.Users {
		users[i] = struct {
			UserId   string `json:"userId"`
			Username string `json:"username"`
		}{
			UserId:   u.UserId,
			Username: u.Username,
		}
	}

	return apigen.PreviewEnvironmentModel{
		AccessSettings: apigen.AccessSettingsModel{
			Users: users,
		},
		ApplicationSettings: apigen.ApplicationSettingsModel{
			Command:              in.Spec.ApplicationSettings.Command,
			EnvironmentVariables: &vars,
			Port:                 in.Spec.ApplicationSettings.Port,
		},
		BuildSettings: apigen.BuildSettings{
			BranchWildcard:       in.Spec.BuildSettings.BranchWildcard,
			BuildAllBranches:     in.Spec.BuildSettings.BuildAllBranches,
			BuildAllPullRequests: in.Spec.BuildSettings.BuildAllPullRequests,
			DockerFilePath:       in.Spec.BuildSettings.DockerfilePath,
		},
		ContainerSettings: apigen.ContainerSettingsModel{
			Registry:   &in.Spec.ContainerRegistry.Registry,
			Repository: &in.Spec.GitSettings.Repository,
		},
		GitSettings: apigen.GitSettingsModel{
			Organization: in.Spec.GitSettings.Organization,
			Repository:   in.Spec.GitSettings.Repository,
		},
		Id:   string(in.GetUID()),
		Name: in.GetName(),
	}
}
