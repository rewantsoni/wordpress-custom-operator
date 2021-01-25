package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	wordpressv1 "wordpress-operator/api/v1"
)

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
