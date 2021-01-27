package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	wordpressv1 "wordpress-operator/api/v1"
)

func createWordPress(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {

	res, err := createWordpressService(r, ctx, log, req, wordpress)
	if err != nil {
		return res, err
	}

	res, err = createPVC(r, ctx, log, req, wordpress, "wp")
	if err != nil {
		return res, err
	}

	res, err = createWordpressDeployment(r, ctx, log, req, wordpress)
	if err != nil {
		return res, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func createWordpressService(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	if objectNotFound(r, ctx, "wordpress", &v1.Service{}, *wordpress) {
		service := newWordpressService(wordpress)
		if err := controllerutil.SetControllerReference(wordpress, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err := r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create Wordpress Service ", "service.name", service.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom Wordpress Service object ", "name", req.NamespacedName.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newWordpressService(wordpress *wordpressv1.Wordpress) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wordpress",
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 80,
				},
			},
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "frontend",
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
}

func createWordpressDeployment(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	if objectNotFound(r, ctx, "wordpress", &appsv1.Deployment{}, *wordpress) {
		deployment := newWordpressDeployment(wordpress)

		if err := controllerutil.SetControllerReference(wordpress, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err := r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to create Wordpress Deployment ", "deployment.name", deployment.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom Wordpress Deployment object ", "name", req.NamespacedName.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newWordpressDeployment(wordpress *wordpressv1.Wordpress) *appsv1.Deployment {

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wordpress",
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "wordpress",
					"tier": "frontend",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "wordpress",
						"tier": "frontend",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "wordpress:4.8-apache",
							Name:  "wordpress",
							Env: []v1.EnvVar{
								{
									Name:  "WORDPRESS_DB_HOST",
									Value: "wordpress-mysql",
								},
								{
									Name: "WORDPRESS_DB_PASSWORD",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "mysql-pass",
											},
											Key: "password",
										},
									},
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "wordpress",
									ContainerPort: 80,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "wordpress-persistent-storage",
									MountPath: "/var/www/html",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "wordpress-persistent-storage",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "wp-pv-claim",
								},
							},
						},
					},
				},
			},
		},
	}
}
