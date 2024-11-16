package server

import (
	"context"
	"os"
	"strconv"

	"github.com/coflnet/pr-env/internal/git"
	"github.com/coflnet/pr-env/internal/keycloak"
	"github.com/coflnet/pr-env/internal/kubeclient"
	apigen "github.com/coflnet/pr-env/internal/server/openapi"
	"github.com/go-logr/logr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	log            *logr.Logger
	kubeClient     *kubeclient.KubeClient
	githubClient   *git.GithubClient
	keycloakClient *keycloak.KeycloakClient
}

func NewServer(ctx context.Context, logger *logr.Logger, githubClient *git.GithubClient, kubeClient *kubeclient.KubeClient, keycloak *keycloak.KeycloakClient) *echo.Echo {
	s := Server{
		githubClient:   githubClient,
		kubeClient:     kubeClient,
		keycloakClient: keycloak,
		log:            logger,
	}

	e := echo.New()

	err := setupAuthenticationMiddleware(ctx)
	if err != nil {
		panic("unable to setup authentication middleware")
	}

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"method", v.Method,
			)

			return nil
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	// authentication routes
	// those are not listed in the openapi spec
	e.Static("/", "internal/server/static")
	e.GET("/login", loginHandler)
	e.GET("/auth/callback", callbackHandler)

	// openapi spec
	e.Static("/api/openapi", staticDir())

	e.GET("/api/github/setupUrl", s.ConfigureInstallation)

	// everything else
	strictServer := apigen.NewStrictHandler(s, []apigen.StrictMiddlewareFunc{})
	apigen.RegisterHandlersWithBaseURL(e, strictServer, "/api/v1")
	return e
}

func MyMiddleware[K ~func(ctx echo.Context, args interface{}) (interface{}, error)](f K, operationID string) K {
	return func(ctx echo.Context, args interface{}) (interface{}, error) {
		// code
		return f(ctx, args)
	}
}

func staticDir() string {
	dir := os.Getenv("STATIC_DIR")
	if dir == "" {
		return "internal/server/openapi"
	}
	return dir
}

func port() int {
	v := os.Getenv("PORT")
	const defaultPort = 8080

	if v == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(v)
	if err != nil {
		return defaultPort
	}

	return p
}

type httpError struct {
	Code     int         `json:"-"`
	Message  interface{} `json:"message"`
	Internal error       `json:"-"` // Stores the error returned by an external dependency
}
