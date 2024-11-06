package controller

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
)

const (
	buildPrefix = "build-"
)

func (r *PreviewEnvironmentInstanceReconciler) rebuildInstance(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	// check if a build version is already available
	builtAvailable := false
	if pei.Status.BuiltVersions != nil {
		for _, version := range pei.Status.BuiltVersions {
			if version.Tag == pei.Spec.CommitHash {
				builtAvailable = true
				break
			}
		}
	}

	if builtAvailable {
		r.log.Info("Built version is already available, skip this build", "namespace", pei.Namespace, "name", pei.Name)
		return nil
	}

	// build the container image
	r.log.Info("Building container image for PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
	err := r.buildContainerImage(ctx, pe, pei)
	if err != nil {
		return err
	}
	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) buildContainerImage(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	job := &kbatch.Job{}
	if err := r.Get(ctx, types.NamespacedName{Name: pei.Name, Namespace: pei.Namespace}, job); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		if job.Status.Active > 0 {
			r.log.Info("Job already running, skipping", "namespace", pei.Namespace, "name", pei.Name)
			return nil
		}

		if job.Status.Succeeded > 0 {
			err = r.Delete(ctx, job)
			if err != nil {
				return err
			}
		}
	}

	pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusBuilding
	if err := r.Status().Update(ctx, pei); err != nil {
		return err
	}

	err := r.buildAndWaitForContainerImage(ctx, pe, pei)
	pei.Status.BuiltVersions = updateBuiltVersions(pei.Status.BuiltVersions, pei.Spec.CommitHash, 10)

	pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusDeploymentOutdated
	if err := r.Status().Update(ctx, pei); err != nil {
		return err
	}

	return err
}

func updateBuiltVersions(versions []coflnetv1alpha1.BuiltVersion, commitHash string, keep int) []coflnetv1alpha1.BuiltVersion {
	if versions == nil {
		versions = []coflnetv1alpha1.BuiltVersion{}
	}

	versions = append(versions, coflnetv1alpha1.BuiltVersion{
		Tag:       commitHash,
		Timestamp: metav1.Now(),
	})

	for len(versions) > keep {
		slices.SortFunc(versions, func(a, b coflnetv1alpha1.BuiltVersion) int {
			return a.Timestamp.Time.Compare(b.Timestamp.Time)
		})

		versions = versions[1:]
	}

	return versions
}

func (r *PreviewEnvironmentInstanceReconciler) buildAndWaitForContainerImage(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	const kanikoSecret = "dockerhub"
	var jobName = fmt.Sprintf("%s%s", buildPrefix, pei.Name)
	var destination = fmt.Sprintf("%s/%s/pr-env:%s-%s-%s", pe.Spec.ContainerRegistry.Registry, pe.Spec.ContainerRegistry.Repository, pei.Spec.GitOrganization, pei.Spec.GitRepository, pei.Spec.CommitHash)

	kanikoJob := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: pei.Namespace,
		},
		Spec: kbatch.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(60),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:v1.23.2",
							Args: []string{
								"--dockerfile=Dockerfile",
								fmt.Sprintf("--context=git://github.com/%s/%s.git#refs/heads/%s", pei.Spec.GitOrganization, pei.Spec.GitRepository, *pei.Spec.Branch),
								fmt.Sprintf("--destination=%s", destination),
								fmt.Sprintf("--custom-platform=%s", "linux/amd64"),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      kanikoSecret,
									MountPath: "/kaniko/.docker",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: kanikoSecret,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: kanikoSecret,
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: "config.json",
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

	var kJob kbatch.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: pei.Namespace}, &kJob); err == nil {
		r.log.Info("Deleting existing kaniko job", "namespace", pei.Namespace, "name", pei.Name)
		if err := r.Delete(ctx, &kJob); err != nil {
			return err
		}

		for {
			if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: pei.Namespace}, &kJob); err != nil {
				if errors.IsNotFound(err) {
					break
				}
				return err
			}
			time.Sleep(time.Second * 1)
		}
	}

	r.log.Info("Creating kaniko job", "namespace", pei.Namespace, "name", pei.Name, "destination", destination)
	if err := r.Create(ctx, kanikoJob); err != nil {
		return err
	}

	counter := 0
	for {
		job := &kbatch.Job{}
		if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: pei.Namespace}, job); err != nil {
			return err
		}

		if job.Status.Succeeded > 0 {
			break
		}

		if job.Status.Failed > 0 {
			err := fmt.Errorf("kaniko job failed")
			r.log.Error(err, "kaniko job failed", "namespace", pei.Namespace, "name", pei.Name)
			return err
		}

		time.Sleep(time.Second * 1)
		if counter > 60*30 {
			err := fmt.Errorf("timeout while waiting for kaniko job to finish")
			r.log.Error(err, "timeout while waiting for kaniko job to finish", "namespace", pei.Namespace, "name", pei.Name)
			return err
		}
	}

	go func() {
		err := r.deleteCompletedPods(ctx, pei)
		if err != nil {
			r.log.Error(err, "Failed to delete completed pods", "namespace", pei.Namespace, "name", pei.Name)
		}
	}()

	return nil
}

func (r *PreviewEnvironmentInstanceReconciler) deleteCompletedPods(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	pods := &corev1.PodList{}
	// PERF: it would be cool to use field selectors here
	// that way we only get the pods we need
	if err := r.List(ctx, pods, &client.ListOptions{
		Namespace: pei.Namespace,
	}); err != nil {
		return err
	}

	// HACK: there has to be a better way to do this
	for _, pod := range pods.Items {
		if !strings.HasPrefix(pod.Name, buildPrefix) {
			continue
		}

		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			statuses := pod.Status.ContainerStatuses
			for _, status := range statuses {
				if status.State.Terminated != nil {
					age := time.Now().Sub(status.State.Terminated.FinishedAt.Time)
					if age > time.Minute*10 {
						if err := r.Delete(ctx, &pod); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
