package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	opts := zap.Options{
		Development: true,
	}
	logger := zap.New(zap.UseFlagOptions(&opts))

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

	e.POST("/webhook", s.GithubWebhookHandler)

	return e
}
