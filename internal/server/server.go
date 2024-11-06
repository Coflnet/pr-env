package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coflnet/pr-env/internal/kubeclient"
	"github.com/coflnet/pr-env/pkg/git"
	"github.com/go-logr/logr"
)

type Server struct {
	log          *logr.Logger
	kubeClient   *kubeclient.KubeClient
	githubClient *git.GithubClient
}

func NewServer(logger *logr.Logger, githubClient *git.GithubClient, kubeClient *kubeclient.KubeClient) *http.Server {
	s := &Server{
		githubClient: githubClient,
		kubeClient:   kubeClient,
		log:          logger,
	}

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", port()),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &server
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
