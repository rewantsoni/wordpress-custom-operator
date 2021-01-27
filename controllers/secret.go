package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	wordpressv1 "wordpress-operator/api/v1"
)

func createSecret(r *WordpressReconciler, ctx context.Context, log logr.Logger, req ctrl.Request, wordpress *wordpressv1.Wordpress) (ctrl.Result, error) {
	if objectNotFound(r, ctx, "mysql-pass", &v1.Secret{}, *wordpress) {

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
