package server

import (
	"context"
	"strconv"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	apigen "github.com/coflnet/pr-env/internal/server/openapi"
	"github.com/labstack/echo/v4"
	"k8s.io/apimachinery/pkg/types"
)

// List all available Environments
// (GET /environment/list)
func (s Server) GetEnvironmentInstanceIdList(ctx context.Context, request apigen.GetEnvironmentInstanceIdListRequestObject) (apigen.GetEnvironmentInstanceIdListResponseObject, error) {
	userId, err := s.userIdFromAuthenticationToken(ctx, request.Params.Authentication)
	if err != nil {
		return nil, echo.NewHTTPError(401, err.Error())
	}

	peis, err := s.kubeClient.ListPreviewEnvironmentInstancesByPreviewEnvironmentId(ctx, userId, types.UID(request.Id))
	if err != nil {
		return nil, echo.NewHTTPError(500, err.Error())
	}

	res := convertToEnvironmentInstanceModelList(*peis)
	return apigen.GetEnvironmentInstanceIdList200JSONResponse(res), nil
}

func convertToEnvironmentInstanceModelList(peis coflnetv1alpha1.PreviewEnvironmentInstanceList) []apigen.PreviewEnvironmentInstanceModel {
	res := make([]apigen.PreviewEnvironmentInstanceModel, 0, len(peis.Items))
	for _, pei := range peis.Items {
		res = append(res, convertToEnvironmentInstanceModel(pei))
	}
	return res
}

func convertToEnvironmentInstanceModel(pei coflnetv1alpha1.PreviewEnvironmentInstance) apigen.PreviewEnvironmentInstanceModel {
	return apigen.PreviewEnvironmentInstanceModel{
		DesiredPhase: pei.Spec.DesiredPhase,
		InstanceGitSettings: apigen.InstanceGitSettingsModel{
			Branch:                pei.Spec.InstanceGitSettings.Branch,
			CommitHash:            &pei.Spec.InstanceGitSettings.CommitHash,
			PullRequestIdentifier: intPtrToStrPtr(pei.Spec.InstanceGitSettings.PullRequestNumber),
		},
		Name:                 pei.GetName(),
		OwnerId:              pei.GetOwner(),
		PreviewEnvironmentId: pei.GetPreviewEnvironmentId(),
	}
}

func intPtrToStrPtr(v *int) *string {
	if v == nil {
		return nil
	}
	return strPtr(strconv.Itoa(*v))
}
