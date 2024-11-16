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
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

var (
	oauthProvider   *oidc.Provider
	oauthConfig     oauth2.Config
	oidcConfigValue oidcConfig
)

func setupAuthenticationMiddleware(ctx context.Context) error {
	var err error
	oidcConfigValue, err = readInAuthMiddlewareConfig()
	if err != nil {
		return err
	}

	oauthProvider, err = oidc.NewProvider(ctx, oidcConfigValue.IssuerUrl)
	if err != nil {
		return err
	}

	oidc.NewProvider(ctx, oidcConfigValue.IssuerUrl)

	oauthConfig = oauth2.Config{
		ClientID:     oidcConfigValue.ClientID,
		ClientSecret: oidcConfigValue.ClientSecret,
		Endpoint:     oauthProvider.Endpoint(),
		RedirectURL:  oidcConfigValue.RedirectUrl,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return nil
}

type oidcConfig struct {
	ClientID     string
	ClientSecret string
	IssuerUrl    string
	RedirectUrl  string
}

func readInAuthMiddlewareConfig() (oidcConfig, error) {
	c := oidcConfig{}
	c.IssuerUrl = os.Getenv("ISSUER_URL")
	if c.IssuerUrl == "" {
		return c, errors.New("ISSUER_URL is not set")
	}

	c.ClientID = os.Getenv("CLIENT_ID")
	if c.ClientID == "" {
		return c, errors.New("CLIENT_ID is not set")
	}

	c.ClientSecret = os.Getenv("CLIENT_SECRET")
	if c.ClientSecret == "" {
		return c, errors.New("CLIENT_SECRET is not set")
	}

	c.RedirectUrl = os.Getenv("REDIRECT_URL")
	if c.RedirectUrl == "" {
		return c, errors.New("REDIRECT_URL is not set")
	}

	return c, nil
}

func (s *Server) userIdFromAuthenticationToken(ctx context.Context, accessToken string) (string, error) {
	// make sure the user is authenticated
	// if not we redirect the user to the login page
	if accessToken == "" {
		return "", fmt.Errorf("empty access token")
	}

	// validate the token
	token, err := oauthProvider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID}).Verify(ctx, accessToken)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("invalid token: %v", err)
	}

	if token.Subject == "" {
		return "", fmt.Errorf("missing user id")
	}

	return token.Subject, nil
}

func loginHandler(c echo.Context) error {
	state, err := randString(16)
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot generate state")
	}
	setCallbackCookie(c, "state", state)

	return c.Redirect(http.StatusFound, oauthConfig.AuthCodeURL(state))
}

func callbackHandler(c echo.Context) error {
	state, err := c.Cookie("state")
	if err != nil {
		return c.String(http.StatusBadRequest, "missing state cookie")
	}
	if c.QueryParam("state") != state.Value {
		return c.String(http.StatusBadRequest, "state mismatch")
	}

	oauth2Token, err := oauthConfig.Exchange(c.Request().Context(), c.QueryParam("code"))
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
