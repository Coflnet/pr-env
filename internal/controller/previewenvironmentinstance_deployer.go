package controller

import (
	"context"
	"fmt"
	"os"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *PreviewEnvironmentInstanceReconciler) redeployInstance(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deploying the environment instance", "namespace", pei.Namespace, "name", pei.Name)

	pei.Status.Phase = coflnetv1alpha1.InstancePhaseDeploying
	err := r.Status().Update(ctx, pei)
	if err != nil {
		return err
	}

	err = r.deployEnvironmentInstance(ctx, pe, pei)
	if err != nil {
		return err
	}
	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployEnvironmentInstance(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	err := r.deployKubernetesDeployment(ctx, pe, pei)
	if err != nil {
		return err
	}

	err = r.deployKubernetesService(ctx, pe, pei)
	if err != nil {
		return err
	}

	err = r.deployAuthenticationProxy(ctx, pe, pei)
	if err != nil {
		r.log.Error(err, "Unable to deploy authentication proxy")
	}

	err = r.deployKubernetesIngress(ctx, pe, pei)
	if err != nil {
		return err
	}

	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deleteResourcesForPreviewEnvironmentInstance(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deleting resources for PreviewEnvironmentInstance", "namespace", pei.GetNamespace(), "name", pei.GetName())

	err := r.deleteKubernetesDeployment(ctx, pei)
	if err != nil {
		r.log.Error(err, "Unable to delete deployment", "namespace", pei.GetNamespace(), "name", pei.GetName())
	}

	err = r.deleteKubernetesService(ctx, pei)
	if err != nil {
		r.log.Error(err, "Unable to delete service", "namespace", pei.GetNamespace(), "name", pei.GetName())
	}

	err = r.deleteKubernetesAuthProxy(ctx, pei)
	if err != nil {
		r.log.Error(err, "Unable to delete auth proxy", "namespace", pei.GetNamespace(), "name", pei.GetName())
	}

	err = r.deleteKubernetesIngress(ctx, pei)
	if err != nil {
		r.log.Error(err, "Unable to delete ingress", "namespace", pei.GetNamespace(), "name", pei.GetName())
	}

	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesDeployment(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	image := coflnetv1alpha1.PreviewEnvironmentInstanceContainerName(pe, pei.BranchOrPullRequestIdentifier(), pei.Spec.InstanceGitSettings.CommitHash)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":   pei.GetName(),
					"owner": pe.GetOwner(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":   pei.GetName(),
						"owner": pe.GetOwner(),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  pei.Name,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(pe.Spec.ApplicationSettings.Port),
									Name:          "http",
								},
							},
							Env: envFromPe(pe),
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	// TODO: reactivate
	// if pe.Spec.ApplicationSettings.Command != nil {
	// 	deployment.Spec.Template.Spec.Containers[0].Command = strings.Split(*pe.Spec.ApplicationSettings.Command, " ")
	// }

	r.log.Info("Check if deployment already exists", "namespace", pei.GetNamespace(), "name", pei.GetName())
	var kDeployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.GetName()}, &kDeployment)
	if err == nil {
		r.log.Info("Deployment already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName())
		err = r.Update(ctx, deployment)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating deployment", "namespace", pei.GetNamespace(), "name", pei.GetName())
	return r.Create(ctx, deployment)
}

func (r *PreviewEnvironmentInstanceReconciler) deleteKubernetesDeployment(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deleting deployment", "namespace", pei.GetNamespace(), "name", pei.GetName())
	err := r.Delete(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesService(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
			Labels: map[string]string{
				"owner": pe.GetOwner(),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": pei.GetName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: int32(pe.Spec.ApplicationSettings.Port),
				},
			},
		},
	}

	r.log.Info("Check if service already exists", "namespace", pei.GetNamespace(), "name", pei.GetName())
	var kService corev1.Service
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.GetName()}, &kService)
	if err == nil {
		r.log.Info("Service already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName())
		err = r.Update(ctx, service)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating service", "namespace", pei.GetNamespace(), "name", pei.GetName())
	return r.Create(ctx, service)
}

func (r *PreviewEnvironmentInstanceReconciler) deleteKubernetesService(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deleting service", "namespace", pei.GetNamespace(), "name", pei.GetName())
	err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// deployAuthenticationProxy deploys a oauth2_proxy instance in front of the application
// with this the application can be protected by keycloak
func (r *PreviewEnvironmentInstanceReconciler) deployAuthenticationProxy(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	err := r.deployAuthenticationProxyDeployment(ctx, pe, pei)
	if err != nil {
		return err
	}

	err = r.deployAuthenticationProxyService(ctx, pe, pei)
	if err != nil {
		return err
	}

	err = r.deployAuthenticationProxyIngress(ctx, pe, pei)
	if err != nil {
		return err
	}

	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deployAuthenticationProxyDeployment(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {

	path := coflnetv1alpha1.PreviewEnvironmentHttpPath(pe, pei)
	redirectUrl := fmt.Sprintf("https://%s%s/oauth2/callback", pe.Spec.ApplicationSettings.IngressHostname, path)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
			Labels: map[string]string{
				"owner": pe.GetOwner(),
				"app":   pei.NameForAuthProxy(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": pei.NameForAuthProxy(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":   pei.NameForAuthProxy(),
						"owner": pe.GetOwner(),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            pei.NameForAuthProxy(),
							Image:           "quay.io/oauth2-proxy/oauth2-proxy:latest",
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4180,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "OAUTH2_PROXY_COOKIE_SECRET",
									Value: "QVlRK3NGYlJBMStYVlkzOUxNVVhyQT09",
								},
							},
							Args: []string{
								"--email-domain=*",
								"--provider=keycloak-oidc",
								fmt.Sprintf("--client-id=%s", authProxyOauthClientId()),
								fmt.Sprintf("--client-secret=%s", authProxyOauthClientSecret()),
								fmt.Sprintf("--redirect-url=%s", redirectUrl),
								fmt.Sprintf("--oidc-issuer-url=%s", authProxyOauthIssuerUrl()),
								"--code-challenge-method=S256",
								"--standard-logging",
								"--auth-logging",
								"--request-logging",
								"--http-address=0.0.0.0:4180",
							},
						},
					},
				},
			},
		},
	}

	r.log.Info("Check if deployment already exists", "namespace", pei.Namespace, "name", pei.Name+"-auth-proxy")
	var kDeployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.NameForAuthProxy()}, &kDeployment)
	if err == nil {
		r.log.Info("Deployment already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName()+"-auth-proxy")
		err = r.Update(ctx, deployment)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating deployment", "namespace", pei.GetNamespace(), "name", pei.NameForAuthProxy())
	return r.Create(ctx, deployment)
}

func (r *PreviewEnvironmentInstanceReconciler) deployAuthenticationProxyService(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
			Labels: map[string]string{
				"app":   pei.NameForAuthProxy(),
				"owner": pe.GetOwner(),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": pei.NameForAuthProxy(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       4180,
					TargetPort: intstr.FromInt(4180),
				},
			},
		},
	}

	r.log.Info("Check if service already exists", "namespace", pei.GetNamespace(), "name", pei.Name+"-auth-proxy")
	var kService corev1.Service
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.NameForAuthProxy()}, &kService)
	if err == nil {
		r.log.Info("Service already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName()+"-auth-proxy")
		err = r.Update(ctx, service)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating service", "namespace", pei.GetNamespace(), "name", pei.NameForAuthProxy())
	return r.Create(ctx, service)
}

func (r *PreviewEnvironmentInstanceReconciler) deployAuthenticationProxyIngress(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	path := fmt.Sprintf("%s/oauth2", coflnetv1alpha1.PreviewEnvironmentHttpPath(pe, pei))
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
			Labels: map[string]string{
				"app":   pei.NameForAuthProxy(),
				"owner": pe.GetOwner(),
			},
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                   "nginx",
				"cert-manager.io/cluster-issuer":                "letsencrypt-prod",
				"nginx.ingress.kubernetes.io/proxy-buffer-size": "512k",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: pe.Spec.ApplicationSettings.IngressHostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: pathPtr(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: pei.GetName(),
											Port: networkingv1.ServiceBackendPort{
												Number: int32(pe.Spec.ApplicationSettings.Port),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{pe.Spec.ApplicationSettings.IngressHostname},
					SecretName: fmt.Sprintf("%s-tls", pei.NameForAuthProxy()),
				},
			},
		},
	}

	r.log.Info("Check if ingress already exists", "namespace", pei.GetNamespace(), "name", pei.GetName()+"-auth-proxy")
	var kIngress networkingv1.Ingress
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.NameForAuthProxy()}, &kIngress)
	if err == nil {
		r.log.Info("Ingress already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName()+"-auth-proxy")
		err = r.Update(ctx, ingress)
		if err != nil {
			return err
		}
		return nil
	}

	r.log.Info("Creating ingress", "namespace", pei.GetNamespace(), "name", pei.GetName()+"-auth-proxy")
	return r.Create(ctx, ingress)
}

func (r *PreviewEnvironmentInstanceReconciler) deleteKubernetesAuthProxy(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	err := r.Delete(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
		},
	})
	if err != nil {
		return err
	}

	r.log.Info("Deleting service", "namespace", pei.GetNamespace(), "name", pei.NameForAuthProxy())
	err = r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
		},
	})
	if err != nil {
		return err
	}

	r.log.Info("Deleting ingress", "namespace", pei.GetNamespace(), "name", pei.NameForAuthProxy())
	return r.Delete(ctx, &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.NameForAuthProxy(),
			Namespace: pei.GetNamespace(),
		},
	})
}

func (r *PreviewEnvironmentInstanceReconciler) deployKubernetesIngress(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	path := coflnetv1alpha1.PreviewEnvironmentHttpPath(pe, pei)

	host := pe.Spec.ApplicationSettings.IngressHostname
	publicEndpoint := fmt.Sprintf("https://%s%s", host, path)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: pe.Spec.ApplicationSettings.IngressHostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: pathPtr(networkingv1.PathTypeImplementationSpecific),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: pei.GetName(),
											Port: networkingv1.ServiceBackendPort{
												Number: int32(pe.Spec.ApplicationSettings.Port),
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

	r.log.Info("Check if ingress already exists", "namespace", pei.GetNamespace(), "name", pei.GetName())
	var kIngress networkingv1.Ingress
	err := r.Get(ctx, client.ObjectKey{Namespace: pei.GetNamespace(), Name: pei.GetName()}, &kIngress)
	if err == nil {
		r.log.Info("Ingress already exists, updating", "namespace", pei.GetNamespace(), "name", pei.GetName())
		err = r.Update(ctx, ingress)
		if err != nil {
			return err
		}
	} else {
		r.log.Info("Creating ingress", "namespace", pei.GetNamespace(), "name", pei.GetName())
		err = r.Create(ctx, ingress)
		if err != nil {
			return err
		}
	}

	pei.Status.PublicFacingUrl = publicEndpoint
	r.log.Info("Updating the status of the PreviewEnvironmentInstance", "namespace", pei.GetNamespace(), "name", pei.GetName(), "publicFacingUrl", publicEndpoint)
	return r.Status().Update(ctx, pei)
}

func (r *PreviewEnvironmentInstanceReconciler) deleteKubernetesIngress(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("Deleting ingress", "namespace", pei.GetNamespace(), "name", pei.GetName())
	err := r.Delete(ctx, &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pei.GetName(),
			Namespace: pei.GetNamespace(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func envFromPe(pe *coflnetv1alpha1.PreviewEnvironment) []corev1.EnvVar {
	res := []corev1.EnvVar{}
	for _, v := range *pe.Spec.ApplicationSettings.EnvironmentVariables {
		res = append(res, corev1.EnvVar{
			Name:  v.Key,
			Value: v.Value,
		})
	}
	return res
}

func pathPtr(s networkingv1.PathType) *networkingv1.PathType {
	return &s
}

func int32Ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func strPtr(s string) *string {
	return &s
}

func authProxyOauthClientId() string {
	return mustReadEnv("AUTH_PROXY_CLIENT_ID")
}

func authProxyOauthClientSecret() string {
	return mustReadEnv("AUTH_PROXY_CLIENT_SECRET")
}

func authProxyOauthIssuerUrl() string {
	return mustReadEnv("AUTH_PROXY_ISSUER_URL")
}

func mustReadEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", key))
	}
	return val
}
