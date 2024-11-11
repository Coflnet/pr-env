package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func newAuthenticationMiddleware(ctx context.Context) (*authenticationMiddleware, error) {
	cfg := &oidcConfig{}
	if err := cfg.readInConfig(); err != nil {
		return nil, err
	}

	oauthProvider, err := oidc.NewProvider(ctx, cfg.IssuerUrl)
	if err != nil {
		return nil, err
	}

	oidc.NewProvider(ctx, cfg.IssuerUrl)
	return &authenticationMiddleware{
		oauthConfig: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     oauthProvider.Endpoint(),
			RedirectURL:  cfg.RedirectUrl,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		oauthProvider: oauthProvider,
		oidcConfig:    *cfg,
	}, nil
}

type authenticationMiddleware struct {
	oauthProvider *oidc.Provider
	oauthConfig   oauth2.Config
	oidcConfig    oidcConfig
}

type oidcConfig struct {
	ClientID     string
	ClientSecret string
	IssuerUrl    string
	RedirectUrl  string
}

func (c *oidcConfig) readInConfig() error {
	c.IssuerUrl = os.Getenv("ISSUER_URL")
	if c.IssuerUrl == "" {
		return errors.New("ISSUER_URL is not set")
	}

	c.ClientID = os.Getenv("CLIENT_ID")
	if c.ClientID == "" {
		return errors.New("CLIENT_ID is not set")
	}

	c.ClientSecret = os.Getenv("CLIENT_SECRET")
	if c.ClientSecret == "" {
		return errors.New("CLIENT_SECRET is not set")
	}

	c.RedirectUrl = os.Getenv("REDIRECT_URL")
	if c.RedirectUrl == "" {
		return errors.New("REDIRECT_URL is not set")
	}

	return nil
}

func (m *authenticationMiddleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if strings.HasPrefix(c.Path(), "/api/openapi") {
			return next(c)
		}

		// make sure the user is authenticated
		// if not we redirect the user to the login page
		authHeaderVal := c.Request().Header.Get("authentication")
		if authHeaderVal == "" {
			authCookie, err := c.Cookie("access_token")
			if err != nil {
				return c.String(http.StatusUnauthorized, "missing access token")
			}

			if authCookie.Value == "" {
				return c.String(http.StatusUnauthorized, "empty access token")
			}
			authHeaderVal = authCookie.Value
		}

		// validate the token
		token, err := m.oauthProvider.Verifier(&oidc.Config{ClientID: m.oauthConfig.ClientID}).Verify(c.Request().Context(), authHeaderVal)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
		}

		if token.Subject == "" {
			return c.String(http.StatusUnauthorized, "missing user id")
		}

		// add the user id to the context
		c.Set("user", token.Subject)

		return next(c)
	}
}

func (s *authenticationMiddleware) loginHandler(c echo.Context) error {
	state, err := randString(16)
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot generate state")
	}
	setCallbackCookie(c, "state", state)

	return c.Redirect(http.StatusFound, s.oauthConfig.AuthCodeURL(state))
}

func (m *authenticationMiddleware) callbackHandler(c echo.Context) error {
	state, err := c.Cookie("state")
	if err != nil {
		return c.String(http.StatusBadRequest, "missing state cookie")
	}
	if c.QueryParam("state") != state.Value {
		return c.String(http.StatusBadRequest, "state mismatch")
	}

	oauth2Token, err := m.oauthConfig.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to exchange token")
	}

	accessTokenCookie := &http.Cookie{
		Name:   "access_token",
		Value:  oauth2Token.AccessToken,
		MaxAge: int(oauth2Token.Expiry.Sub(time.Now()).Seconds()),
		Secure: c.IsTLS(),
		Path:   "/",
	}
	c.SetCookie(accessTokenCookie)

	refreshTokenCookie := &http.Cookie{
		Name:   "refresh_token",
		Value:  oauth2Token.RefreshToken,
		MaxAge: int(time.Hour.Hours()),
		Secure: c.IsTLS(),
		Path:   "/",
	}
	c.SetCookie(refreshTokenCookie)

	return c.Redirect(http.StatusFound, "/")
}

func setCallbackCookie(c echo.Context, name, value string) {

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   c.IsTLS(),
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
