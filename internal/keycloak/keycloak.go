package keycloak

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"
	_ "github.com/joho/godotenv"
)

type KeycloakClient struct {
	log logr.Logger
}

func NewKeycloakClient(logger logr.Logger) *KeycloakClient {
	return &KeycloakClient{
		log: logger,
	}
}

func (k *KeycloakClient) client() *gocloak.GoCloak {
	url := os.Getenv("KEYCLOAK_URL")
	return gocloak.NewClient(url)
}

func (k *KeycloakClient) adminToken(client *gocloak.GoCloak) *gocloak.JWT {
	user := os.Getenv("KEYCLOAK_USERNAME")
	pass := os.Getenv("KEYCLOAK_PASSWORD")
	realm := os.Getenv("KEYCLOAK_REALM")

	token, err := client.LoginAdmin(context.TODO(), user, pass, realm)
	if err != nil {
		panic("Something wrong with the credentials or url")
	}

	return token
}

func (k *KeycloakClient) UserInformation(ctx context.Context, id string) (*gocloak.User, error) {
	c := k.client()
	user, err := c.GetUserByID(ctx, k.adminToken(c).AccessToken, realm(), id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (k *KeycloakClient) GithubInstallationIdForUser(ctx context.Context, id string) (int, error) {
	user, err := k.UserInformation(ctx, id)
	if err != nil {
		return 0, err
	}

	if user.Attributes == nil {
		return 0, InstallationIdDoesNotExistError{
			UserId: id,
		}
	}

	attr, ok := (*user.Attributes)["githubInstallationId"]
	if !ok || len(attr) == 0 {
		return 0, InstallationIdDoesNotExistError{
			UserId: id,
		}
	}

	installationId, err := strconv.Atoi(attr[0])
	if err != nil {
		return 0, err
	}

	return installationId, nil
}

func (k *KeycloakClient) UserByGithubId(ctx context.Context, id int) (*gocloak.User, error) {
	c := k.client()
	t := k.adminToken(c)

	users, err := c.GetUsers(ctx, t.AccessToken, realm(), gocloak.GetUsersParams{
		IDPUserID: strPtr(strconv.Itoa(id)),
		IDPAlias:  strPtr("github"),
	})

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, UserNotFound{
			ID: id,
		}
	}

	return users[0], nil
}

func (k *KeycloakClient) SetGithubInstallationIdForUser(ctx context.Context, userId string, installationId int) error {
	c := k.client()
	t := k.adminToken(c)

	user, err := k.UserInformation(ctx, userId)
	if err != nil {
		return err
	}

	if user.Attributes == nil {
		user.Attributes = &map[string][]string{}
	}

	(*user.Attributes)["githubInstallationId"] = []string{strconv.Itoa(installationId)}

	k.log.Info("Updating user in keycloak", "id", userId, "installationId", installationId)
	return c.UpdateUser(context.TODO(), t.AccessToken, realm(), *user)
}

type UserNotFound struct {
	ID int
}

func (u UserNotFound) Error() string {
	return fmt.Sprintf("User with github id %d not found", u.ID)
}

type InstallationIdDoesNotExistError struct {
	UserId string
}

func (i InstallationIdDoesNotExistError) Error() string {
	return fmt.Sprintf("User %s does not have an installation id", i.UserId)
}

func realm() string {
	v := os.Getenv("KEYCLOAK_REALM")
	if v == "" {
		panic("KEYCLOAK_REALM is not set")
	}
	return v
}

func strPtr(i string) *string {
	return &i
}
