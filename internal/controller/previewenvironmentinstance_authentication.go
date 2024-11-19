package controller

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
)

func (r *PreviewEnvironmentInstanceReconciler) setupAuthenticationForInstance(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	err := r.createKeycloakGroupIfNotExists(ctx, pei)
	if err != nil {
		return err
	}

	return r.addUsersToGroup(ctx, pei)
}

func (r *PreviewEnvironmentInstanceReconciler) createKeycloakGroupIfNotExists(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	group, err := r.keycloakClient.GroupByName(ctx, pei.GetName())
	if err != nil {
		return err
	}

	// group already exists
	if group != nil {
		r.log.Info("Group already exists", "group", *group.Name, "instance", pei.GetName())
		return nil
	}

	g, err := r.keycloakClient.CreateGroup(ctx, pei.GetName())
	if err != nil {
		return err
	}

	r.log.Info("Group created", "group", *g.Name, "instance", pei.GetName())
	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) addUsersToGroup(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	// add the owner to the group
	ownerId := pei.GetOwner()
	r.log.Info("Adding owner to group", "owner", ownerId, "group", pei.GetName())
	err := r.keycloakClient.AddUserToGroup(ctx, ownerId, pei.GetName())
	if err != nil {
		return err
	}

	// add the users to the group
	r.log.Info("Adding users to group is not implemented yet", "group", pei.GetName())
	return nil
}
