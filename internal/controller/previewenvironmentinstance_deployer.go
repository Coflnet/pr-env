package controller

import (
	"context"
	"fmt"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *PreviewEnvironmentInstanceReconciler) redeployInstance(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deploying the environment instance", "namespace", pei.Namespace, "name", pei.Name)

	err := r.deployEnvironmentInstance(ctx, pei)
	if err != nil {
		return err
	}
	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployEnvironmentInstance(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusDeploying
	if err := r.Status().Update(ctx, pei); err != nil {
		return err
	}

	err := r.deployKubernetesDeployment(ctx, pei)
	if err != nil {
		return err
	}

	err = r.deployKubernetesService(ctx, pei)
	if err != nil {
		return err
	}

	err = r.deployKubernetesIngress(ctx, pei)
	if err != nil {
		return err
	}

	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesDeployment(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.Name,
			Namespace: pei.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": pei.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": pei.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  pei.Name,
							Image: fmt.Sprintf("muehlhansfl/pr-env:%s-%s-%s", pei.Spec.GitOrganization, pei.Spec.GitRepository, pei.Spec.CommitHash),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
								},
							},
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	r.log.Info("Check if deployment already exists", "namespace", pei.Namespace, "name", pei.Name)
	var kDeployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.Namespace, Name: pei.Name}, &kDeployment)
	if err == nil {
		r.log.Info("Deployment already exists, updating", "namespace", pei.Namespace, "name", pei.Name)
		err = r.Update(ctx, deployment)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating deployment", "namespace", pei.Namespace, "name", pei.Name)
	return r.Create(ctx, deployment)
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesService(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.Name,
			Namespace: pei.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": pei.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}

	r.log.Info("Check if service already exists", "namespace", pei.Namespace, "name", pei.Name)
	var kService corev1.Service
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.Namespace, Name: pei.Name}, &kService)
	if err == nil {
		r.log.Info("Service already exists, updating", "namespace", pei.Namespace, "name", pei.Name)
		err = r.Update(ctx, service)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating service", "namespace", pei.Namespace, "name", pei.Name)
	return r.Create(ctx, service)
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesIngress(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.Name,
			Namespace: pei.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("pr-env-%s", pei.Name),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: strPtr(networkingv1.PathTypeImplementationSpecific),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: pei.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	r.log.Info("Check if ingress already exists", "namespace", pei.Namespace, "name", pei.Name)
	var kIngress networkingv1.Ingress
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.Namespace, Name: pei.Name}, &kIngress)
	if err == nil {
		r.log.Info("Ingress already exists, updating", "namespace", pei.Namespace, "name", pei.Name)
		err = r.Update(ctx, ingress)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating ingress", "namespace", pei.Namespace, "name", pei.Name)
	return r.Create(ctx, ingress)
}

func strPtr(s networkingv1.PathType) *networkingv1.PathType {
	return &s
}

func int32Ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}
