// Package apigen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package apigen

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
)

// ServerApplicationSettingsModel defines model for server.applicationSettingsModel.
type ServerApplicationSettingsModel struct {
	Hostname *string `json:"hostname,omitempty"`
	Port     *int    `json:"port,omitempty"`
}

// ServerContainerSettingsModel defines model for server.containerSettingsModel.
type ServerContainerSettingsModel struct {
	Image    *string `json:"image,omitempty"`
	Registry *string `json:"registry,omitempty"`
}

// ServerEnvironmentModel defines model for server.environmentModel.
type ServerEnvironmentModel struct {
	ApplicationSettings *ServerApplicationSettingsModel `json:"applicationSettings,omitempty"`
	ContainerSettings   *ServerContainerSettingsModel   `json:"containerSettings,omitempty"`
	GitSettings         *ServerGitSettingsModel         `json:"gitSettings,omitempty"`
	Id                  *string                         `json:"id,omitempty"`
	Name                *string                         `json:"name,omitempty"`
}

// ServerGitSettingsModel defines model for server.gitSettingsModel.
type ServerGitSettingsModel struct {
	Organization *string `json:"organization,omitempty"`
	Repository   *string `json:"repository,omitempty"`
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

// PostEnvironmentJSONRequestBody defines body for PostEnvironment for application/json ContentType.
type PostEnvironmentJSONRequestBody = ServerEnvironmentModel

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

}
