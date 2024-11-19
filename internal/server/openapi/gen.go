// Package apigen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package apigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
)

// ApplicationSettingsModel defines model for applicationSettingsModel.
type ApplicationSettingsModel struct {
	Command              *string                     `json:"command,omitempty"`
	EnvironmentVariables *[]EnvironmentVariableModel `json:"environmentVariables,omitempty"`
	Port                 int                         `json:"port"`
}

// BuildSettings defines model for buildSettings.
type BuildSettings struct {
	BranchWildcard       *string `json:"branchWildcard,omitempty"`
	BuildAllBranches     bool    `json:"buildAllBranches"`
	BuildAllPullRequests bool    `json:"buildAllPullRequests"`
	DockerFilePath       *string `json:"dockerFilePath,omitempty"`
}

// ContainerSettingsModel defines model for containerSettingsModel.
type ContainerSettingsModel struct {
	Registry   *string `json:"registry,omitempty"`
	Repository *string `json:"repository,omitempty"`
}

// EnvironmentVariableModel defines model for environmentVariableModel.
type EnvironmentVariableModel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GitSettingsModel defines model for gitSettingsModel.
type GitSettingsModel struct {
	Organization string `json:"organization"`
	Repository   string `json:"repository"`
}

// GithubRepositoryModel defines model for githubRepositoryModel.
type GithubRepositoryModel struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// InstanceGitSettingsModel defines model for instanceGitSettingsModel.
type InstanceGitSettingsModel struct {
	Branch                *string `json:"branch,omitempty"`
	CommitHash            *string `json:"commitHash,omitempty"`
	PullRequestIdentifier *string `json:"pullRequestIdentifier,omitempty"`
}

// PreviewEnvironmentInstanceModel defines model for previewEnvironmentInstanceModel.
type PreviewEnvironmentInstanceModel struct {
	CurrentPhase         string                   `json:"currentPhase"`
	DesiredPhase         string                   `json:"desiredPhase"`
	InstanceGitSettings  InstanceGitSettingsModel `json:"instanceGitSettings"`
	Name                 string                   `json:"name"`
	OwnerId              string                   `json:"ownerId"`
	PreviewEnvironmentId string                   `json:"previewEnvironmentId"`
	PublicFacingUrl      *string                  `json:"publicFacingUrl,omitempty"`
}

// PreviewEnvironmentModel defines model for previewEnvironmentModel.
type PreviewEnvironmentModel struct {
	ApplicationSettings ApplicationSettingsModel `json:"applicationSettings"`
	BuildSettings       BuildSettings            `json:"buildSettings"`
	ContainerSettings   ContainerSettingsModel   `json:"containerSettings"`
	GitSettings         GitSettingsModel         `json:"gitSettings"`
	Id                  string                   `json:"id"`
	Name                string                   `json:"name"`
}

// ServerHttpError defines model for server.httpError.
type ServerHttpError struct {
	Message *map[string]interface{} `json:"message,omitempty"`
}

// PostEnvironmentParams defines parameters for PostEnvironment.
type PostEnvironmentParams struct {
	// Authentication Authentication token
	Authentication string `json:"authentication"`
}

// GetEnvironmentInstanceIdListParams defines parameters for GetEnvironmentInstanceIdList.
type GetEnvironmentInstanceIdListParams struct {
	// Authentication Authentication token
	Authentication string `json:"authentication"`
}

// GetEnvironmentListParams defines parameters for GetEnvironmentList.
type GetEnvironmentListParams struct {
	// Authentication Authentication token
	Authentication string `json:"authentication"`
}

// DeleteEnvironmentIdParams defines parameters for DeleteEnvironmentId.
type DeleteEnvironmentIdParams struct {
	// Authentication Authentication token
	Authentication string `json:"authentication"`
}

// GetGithubRepositoriesParams defines parameters for GetGithubRepositories.
type GetGithubRepositoriesParams struct {
	// Authentication Authentication token
	Authentication string `json:"authentication"`
}

// PostEnvironmentJSONRequestBody defines body for PostEnvironment for application/json ContentType.
type PostEnvironmentJSONRequestBody = PreviewEnvironmentModel

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Creates a new environment
	// (POST /environment)
	PostEnvironment(ctx echo.Context, params PostEnvironmentParams) error
	// Lists all instances of an environment
	// (GET /environment-instance/{id}/list)
	GetEnvironmentInstanceIdList(ctx echo.Context, id string, params GetEnvironmentInstanceIdListParams) error
	// List all available Environments
	// (GET /environment/list)
	GetEnvironmentList(ctx echo.Context, params GetEnvironmentListParams) error
	// Deletes an environment
	// (DELETE /environment/{id})
	DeleteEnvironmentId(ctx echo.Context, id string, params DeleteEnvironmentIdParams) error
	// Lists all the repositories of the authenticated user
	// (GET /github/repositories)
	GetGithubRepositories(ctx echo.Context, params GetGithubRepositoriesParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// PostEnvironment converts echo context to params.
func (w *ServerInterfaceWrapper) PostEnvironment(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params PostEnvironmentParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "authentication" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("authentication")]; found {
		var Authentication string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for authentication, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "authentication", valueList[0], &Authentication, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authentication: %s", err))
		}

		params.Authentication = Authentication
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter authentication is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostEnvironment(ctx, params)
	return err
}

// GetEnvironmentInstanceIdList converts echo context to params.
func (w *ServerInterfaceWrapper) GetEnvironmentInstanceIdList(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithOptions("simple", "id", ctx.Param("id"), &id, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetEnvironmentInstanceIdListParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "authentication" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("authentication")]; found {
		var Authentication string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for authentication, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "authentication", valueList[0], &Authentication, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authentication: %s", err))
		}

		params.Authentication = Authentication
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter authentication is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetEnvironmentInstanceIdList(ctx, id, params)
	return err
}

// GetEnvironmentList converts echo context to params.
func (w *ServerInterfaceWrapper) GetEnvironmentList(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetEnvironmentListParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "authentication" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("authentication")]; found {
		var Authentication string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for authentication, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "authentication", valueList[0], &Authentication, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authentication: %s", err))
		}

		params.Authentication = Authentication
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter authentication is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetEnvironmentList(ctx, params)
	return err
}

// DeleteEnvironmentId converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteEnvironmentId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithOptions("simple", "id", ctx.Param("id"), &id, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params DeleteEnvironmentIdParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "authentication" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("authentication")]; found {
		var Authentication string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for authentication, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "authentication", valueList[0], &Authentication, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authentication: %s", err))
		}

		params.Authentication = Authentication
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter authentication is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.DeleteEnvironmentId(ctx, id, params)
	return err
}

// GetGithubRepositories converts echo context to params.
func (w *ServerInterfaceWrapper) GetGithubRepositories(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetGithubRepositoriesParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "authentication" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("authentication")]; found {
		var Authentication string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for authentication, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "authentication", valueList[0], &Authentication, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authentication: %s", err))
		}

		params.Authentication = Authentication
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter authentication is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetGithubRepositories(ctx, params)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/environment", wrapper.PostEnvironment)
	router.GET(baseURL+"/environment-instance/:id/list", wrapper.GetEnvironmentInstanceIdList)
	router.GET(baseURL+"/environment/list", wrapper.GetEnvironmentList)
	router.DELETE(baseURL+"/environment/:id", wrapper.DeleteEnvironmentId)
	router.GET(baseURL+"/github/repositories", wrapper.GetGithubRepositories)

}

type PostEnvironmentRequestObject struct {
	Params PostEnvironmentParams
	Body   *PostEnvironmentJSONRequestBody
}

type PostEnvironmentResponseObject interface {
	VisitPostEnvironmentResponse(w http.ResponseWriter) error
}

type PostEnvironment200JSONResponse PreviewEnvironmentModel

func (response PostEnvironment200JSONResponse) VisitPostEnvironmentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type PostEnvironment400JSONResponse ServerHttpError

func (response PostEnvironment400JSONResponse) VisitPostEnvironmentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type PostEnvironment409JSONResponse ServerHttpError

func (response PostEnvironment409JSONResponse) VisitPostEnvironmentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(409)

	return json.NewEncoder(w).Encode(response)
}

type PostEnvironment500JSONResponse ServerHttpError

func (response PostEnvironment500JSONResponse) VisitPostEnvironmentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetEnvironmentInstanceIdListRequestObject struct {
	Id     string `json:"id"`
	Params GetEnvironmentInstanceIdListParams
}

type GetEnvironmentInstanceIdListResponseObject interface {
	VisitGetEnvironmentInstanceIdListResponse(w http.ResponseWriter) error
}

type GetEnvironmentInstanceIdList200JSONResponse []PreviewEnvironmentInstanceModel

func (response GetEnvironmentInstanceIdList200JSONResponse) VisitGetEnvironmentInstanceIdListResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetEnvironmentInstanceIdList401JSONResponse ServerHttpError

func (response GetEnvironmentInstanceIdList401JSONResponse) VisitGetEnvironmentInstanceIdListResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type GetEnvironmentInstanceIdList500JSONResponse ServerHttpError

func (response GetEnvironmentInstanceIdList500JSONResponse) VisitGetEnvironmentInstanceIdListResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetEnvironmentListRequestObject struct {
	Params GetEnvironmentListParams
}

type GetEnvironmentListResponseObject interface {
	VisitGetEnvironmentListResponse(w http.ResponseWriter) error
}

type GetEnvironmentList200JSONResponse []PreviewEnvironmentModel

func (response GetEnvironmentList200JSONResponse) VisitGetEnvironmentListResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetEnvironmentList500JSONResponse ServerHttpError

func (response GetEnvironmentList500JSONResponse) VisitGetEnvironmentListResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type DeleteEnvironmentIdRequestObject struct {
	Id     string `json:"id"`
	Params DeleteEnvironmentIdParams
}

type DeleteEnvironmentIdResponseObject interface {
	VisitDeleteEnvironmentIdResponse(w http.ResponseWriter) error
}

type DeleteEnvironmentId200JSONResponse PreviewEnvironmentModel

func (response DeleteEnvironmentId200JSONResponse) VisitDeleteEnvironmentIdResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type DeleteEnvironmentId400JSONResponse ServerHttpError

func (response DeleteEnvironmentId400JSONResponse) VisitDeleteEnvironmentIdResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type DeleteEnvironmentId404JSONResponse ServerHttpError

func (response DeleteEnvironmentId404JSONResponse) VisitDeleteEnvironmentIdResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response)
}

type DeleteEnvironmentId500JSONResponse ServerHttpError

func (response DeleteEnvironmentId500JSONResponse) VisitDeleteEnvironmentIdResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetGithubRepositoriesRequestObject struct {
	Params GetGithubRepositoriesParams
}

type GetGithubRepositoriesResponseObject interface {
	VisitGetGithubRepositoriesResponse(w http.ResponseWriter) error
}

type GetGithubRepositories200JSONResponse []GithubRepositoryModel

func (response GetGithubRepositories200JSONResponse) VisitGetGithubRepositoriesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetGithubRepositories401JSONResponse ServerHttpError

func (response GetGithubRepositories401JSONResponse) VisitGetGithubRepositoriesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type GetGithubRepositories500JSONResponse ServerHttpError

func (response GetGithubRepositories500JSONResponse) VisitGetGithubRepositoriesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Creates a new environment
	// (POST /environment)
	PostEnvironment(ctx context.Context, request PostEnvironmentRequestObject) (PostEnvironmentResponseObject, error)
	// Lists all instances of an environment
	// (GET /environment-instance/{id}/list)
	GetEnvironmentInstanceIdList(ctx context.Context, request GetEnvironmentInstanceIdListRequestObject) (GetEnvironmentInstanceIdListResponseObject, error)
	// List all available Environments
	// (GET /environment/list)
	GetEnvironmentList(ctx context.Context, request GetEnvironmentListRequestObject) (GetEnvironmentListResponseObject, error)
	// Deletes an environment
	// (DELETE /environment/{id})
	DeleteEnvironmentId(ctx context.Context, request DeleteEnvironmentIdRequestObject) (DeleteEnvironmentIdResponseObject, error)
	// Lists all the repositories of the authenticated user
	// (GET /github/repositories)
	GetGithubRepositories(ctx context.Context, request GetGithubRepositoriesRequestObject) (GetGithubRepositoriesResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// PostEnvironment operation middleware
func (sh *strictHandler) PostEnvironment(ctx echo.Context, params PostEnvironmentParams) error {
	var request PostEnvironmentRequestObject

	request.Params = params

	var body PostEnvironmentJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.PostEnvironment(ctx.Request().Context(), request.(PostEnvironmentRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostEnvironment")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(PostEnvironmentResponseObject); ok {
		return validResponse.VisitPostEnvironmentResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetEnvironmentInstanceIdList operation middleware
func (sh *strictHandler) GetEnvironmentInstanceIdList(ctx echo.Context, id string, params GetEnvironmentInstanceIdListParams) error {
	var request GetEnvironmentInstanceIdListRequestObject

	request.Id = id
	request.Params = params

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetEnvironmentInstanceIdList(ctx.Request().Context(), request.(GetEnvironmentInstanceIdListRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetEnvironmentInstanceIdList")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetEnvironmentInstanceIdListResponseObject); ok {
		return validResponse.VisitGetEnvironmentInstanceIdListResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetEnvironmentList operation middleware
func (sh *strictHandler) GetEnvironmentList(ctx echo.Context, params GetEnvironmentListParams) error {
	var request GetEnvironmentListRequestObject

	request.Params = params

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetEnvironmentList(ctx.Request().Context(), request.(GetEnvironmentListRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetEnvironmentList")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetEnvironmentListResponseObject); ok {
		return validResponse.VisitGetEnvironmentListResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// DeleteEnvironmentId operation middleware
func (sh *strictHandler) DeleteEnvironmentId(ctx echo.Context, id string, params DeleteEnvironmentIdParams) error {
	var request DeleteEnvironmentIdRequestObject

	request.Id = id
	request.Params = params

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.DeleteEnvironmentId(ctx.Request().Context(), request.(DeleteEnvironmentIdRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "DeleteEnvironmentId")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(DeleteEnvironmentIdResponseObject); ok {
		return validResponse.VisitDeleteEnvironmentIdResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetGithubRepositories operation middleware
func (sh *strictHandler) GetGithubRepositories(ctx echo.Context, params GetGithubRepositoriesParams) error {
	var request GetGithubRepositoriesRequestObject

	request.Params = params

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetGithubRepositories(ctx.Request().Context(), request.(GetGithubRepositoriesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetGithubRepositories")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetGithubRepositoriesResponseObject); ok {
		return validResponse.VisitGetGithubRepositoriesResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}
