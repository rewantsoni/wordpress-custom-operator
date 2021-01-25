package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	wordpressv1 "wordpress-operator/api/v1"
)

func createMySQL(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	res, err := createMySQLService(r, ctx, log, req, wordpress, "wordpress-mysql")
	if err != nil {
		return res, err
	}

	res, err = createPVC(r, ctx, log, req, wordpress, "mysql")
	if err != nil {
		return res, err
	}

	res, err = createMySQLDeployment(r, ctx, log, req, wordpress, "wordpress-mysql")
	if err != nil {
		return res, err
	}

	return ctrl.Result{Requeue: true}, nil

}

func createMySQLService(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      "wordpress-mysql",
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &v1.Service{})
	if err != nil && errors.IsNotFound(err) {
		service := newMySQLService(wordpress)
		if err := controllerutil.SetControllerReference(wordpress, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err := r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create Service "+name, "service.name", service.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom Service object "+name, "name", req.NamespacedName.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newMySQLService(wordpress *wordpressv1.Wordpress) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wordpress-mysql",
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 3306,
				},
			},
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "mysql",
			},
			ClusterIP: "None",
		},
	}
}

func createMySQLDeployment(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      name,
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &appsv1.Deployment{})
	if err != nil && errors.IsNotFound(err) {
		deployment := newMySQLDeployment(wordpress, name)

		if err := controllerutil.SetControllerReference(wordpress, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err := r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to create Deployment "+name, "deployment.name", deployment.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom Deployment object "+name, "name", req.NamespacedName.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newMySQLDeployment(wordpress *wordpressv1.Wordpress, name string) *appsv1.Deployment {

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "wordpress",
					"tier": "mysql",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "wordpress",
						"tier": "mysql",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "mysql:5.6",
							Name:  "mysql",
							Env: []v1.EnvVar{
								{
									Name: "MYSQL_ROOT_PASSWORD",
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
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{IntVal: 3306},
									},
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "mysql",
									ContainerPort: 3306,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "mysql-persistent-storage",
									MountPath: "/var/lib/mysql",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "mysql-persistent-storage",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mysql-pv-claim",
								},
							},
						},
					},
				},
			},
		},
	}
}
