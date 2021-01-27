package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	wordpressv1 "wordpress-operator/api/v1"
)

func objectNotFound(r *WordpressReconciler, ctx context.Context, key string, obj client.Object, wordpress wordpressv1.Wordpress) bool {
	toFind := types.NamespacedName{
		Name:      key,
		Namespace: wordpress.Namespace,
	}

	err := r.Get(ctx, toFind, obj)
	return err != nil && errors.IsNotFound(err)
}
