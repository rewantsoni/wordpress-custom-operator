/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wordpressv1 "wordpress-operator/api/v1"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=Deployment,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=Service,verbs=get;list;watch;create;update;patch;deleted

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wordpress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *WordpressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx = context.Background()
	log := r.Log.WithValues("wordpress", req.NamespacedName)

	wordpress := &wordpressv1.Wordpress{}
	err := r.Get(ctx, req.NamespacedName, wordpress)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	res, err := createSecret(r, ctx, log, req, wordpress)
	if err != nil {
		return res, err
	}

	//res, err = createPV(r, ctx, log, req, wordpress)
	//if err != nil {
	//	return res, err
	//}

	res, err = createWordPress(r, ctx, log, req, wordpress)
	if err != nil {
		return res, err
	}

	res, err = createMySQL(r, ctx, log, req, wordpress)
	if err != nil {
		return res, err
	}
	// your logic here

	return ctrl.Result{Requeue: true}, nil
}

func createMySQL(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	res, err := createServiceMySQL(r, ctx, log, req, wordpress, "wordpress-mysql")
	if err != nil {
		return res, err
	}

	res, err = createPVC(r, ctx, log, req, wordpress, "mysql")
	if err != nil {
		return res, err
	}

	res, err = createDeploymentMySQL(r, ctx, log, req, wordpress, "wordpress-mysql")
	if err != nil {
		return res, err
	}

	return ctrl.Result{Requeue: true}, nil

}

func createWordPress(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {

	res, err := createService(r, ctx, log, req, wordpress, "wordpress")
	if err != nil {
		return res, err
	}

	res, err = createPVC(r, ctx, log, req, wordpress, "wp")
	if err != nil {
		return res, err
	}

	res, err = createDeployment(r, ctx, log, req, wordpress, "wordpress")
	if err != nil {
		return res, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func createPVC(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      name + "-pv-claim",
		Namespace: wordpress.Namespace,
	}

	err := r.Get(ctx, toFind, &v1.PersistentVolumeClaim{})
	if err != nil && errors.IsNotFound(err) {
		pvc := newPVC(wordpress, name)

		if err = controllerutil.SetControllerReference(wordpress, pvc, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err := r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create PVC "+name, "pvc.name", pvc.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom PVC object "+name, "name", req.NamespacedName.Name)

		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newPVC(wordpress *wordpressv1.Wordpress, name string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-pv-claim",
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}
}

func createDeploymentMySQL(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      name,
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &appsv1.Deployment{})
	if err != nil && errors.IsNotFound(err) {
		deployment := newDeploymentMySQL(wordpress, name)

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

func newDeploymentMySQL(wordpress *wordpressv1.Wordpress, name string) *appsv1.Deployment {

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

func createDeployment(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      name,
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &appsv1.Deployment{})
	if err != nil && errors.IsNotFound(err) {
		deployment := newDeployment(wordpress, name)

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

func newDeployment(wordpress *wordpressv1.Wordpress, name string) *appsv1.Deployment {

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  name,
					"tier": "frontend",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  name,
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
									Name:          name,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      name + "-persistent-storage",
									MountPath: "/var/www/html",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: name + "-persistent-storage",
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

func createServiceMySQL(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      "wordpress-mysql",
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &v1.Service{})
	if err != nil && errors.IsNotFound(err) {
		service := newServiceMySQL(wordpress)
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

func newServiceMySQL(wordpress *wordpressv1.Wordpress) *v1.Service {
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

func createService(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress, name string) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      name,
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &v1.Service{})
	if err != nil && errors.IsNotFound(err) {
		service := newService(wordpress, name)
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

func newService(wordpress *wordpressv1.Wordpress, name string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wordpress.Namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 80,
				},
			},
			Selector: map[string]string{
				"app":  name,
				"tier": "frontend",
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
}

func createSecret(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	toFind := types.NamespacedName{
		Name:      "mysql-pass",
		Namespace: wordpress.Namespace,
	}
	err := r.Get(ctx, toFind, &v1.Secret{})
	if err != nil && errors.IsNotFound(err) {

		secret := newSecret(wordpress)

		if err := controllerutil.SetControllerReference(wordpress, secret, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err := r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Failed to create Secret", "secret.name", secret.Name)
			return ctrl.Result{}, err
		}
		log.Info("Returned custom secret object", "name", req.NamespacedName.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func newSecret(wordpress *wordpressv1.Wordpress) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-pass",
			Namespace: wordpress.Namespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"password": []byte(wordpress.Spec.SqlRootPassword),
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wordpressv1.Wordpress{}).
		Complete(r)
}
