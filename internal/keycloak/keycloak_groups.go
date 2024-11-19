package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

func (k *KeycloakClient) GroupByName(ctx context.Context, name string) (*gocloak.Group, error) {
	c := k.client()
	token := k.adminToken(c)

	groups, err := c.GetGroups(ctx, token.AccessToken, realm(), gocloak.GetGroupsParams{
		Search: &name,
	})
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, nil
	}

	return groups[0], nil
}

func (k *KeycloakClient) CreateGroup(ctx context.Context, groupName string) (*gocloak.Group, error) {
	c := k.client()
	token := k.adminToken(c)

	group := gocloak.Group{
		Name: &groupName,
	}

	_, err := c.CreateGroup(ctx, token.AccessToken, realm(), group)
	if err != nil {
		return nil, err
	}

	return k.GroupByName(ctx, groupName)
}

func (k *KeycloakClient) AddUserToGroup(ctx context.Context, userId string, groupName string) error {
	c := k.client()
	token := k.adminToken(c)

	group, err := k.GroupByName(ctx, groupName)
	if err != nil {
		return err
	}

	if group == nil {
		return fmt.Errorf("group %s does not exist", groupName)
	}

	err = c.AddUserToGroup(ctx, token.AccessToken, realm(), userId, *group.ID)
	if err != nil {
		return err
	}

	return nil
}
