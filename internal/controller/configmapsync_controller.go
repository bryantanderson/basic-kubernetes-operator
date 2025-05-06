/*
Copyright 2025.

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

package controller

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/bryantanderson/basic-kubernetes-operator/api/v1"
)

// ConfigMapSyncReconciler reconciles a ConfigMapSync object
type ConfigMapSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.bryantanderson.github.io,resources=configmapsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.bryantanderson.github.io,resources=configmapsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.bryantanderson.github.io,resources=configmapsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ConfigMapSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the ConfigMapSync instance
	configMapSync := &appsv1.ConfigMapSync{}
	if err := r.Get(ctx, req.NamespacedName, configMapSync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the source ConfigMap
	sourceConfigMap := &corev1.ConfigMap{}
	sourceConfigMapName := types.NamespacedName{
		Namespace: configMapSync.Spec.SourceNamespace,
		Name:      configMapSync.Spec.ConfigMapName,
	}
	if err := r.Get(ctx, sourceConfigMapName, sourceConfigMap); err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the destination ConfigMap
	destinationConfigMap := &corev1.ConfigMap{}
	destinationConfigMapName := types.NamespacedName{
		Namespace: configMapSync.Spec.DestinationNamespace,
		Name:      configMapSync.Spec.ConfigMapName,
	}
	if err := r.Get(ctx, destinationConfigMapName, destinationConfigMap); err != nil {
		// If the err is not a basic NotFound error, return
		if !strings.Contains(err.Error(), "not found") {
			return ctrl.Result{}, err
		}
		// Otherwise simply create a new ConfigMap within the destination namespace,
		// which is just a copy of the source ConfigMap
		destinationConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapSync.Spec.ConfigMapName,
				Namespace: configMapSync.Spec.DestinationNamespace,
			},
			Data: sourceConfigMap.Data,
		}
		if err := r.Create(ctx, destinationConfigMap); err != nil {
			return ctrl.Result{}, err
		}
	}

	// If we've reached this point then both config maps exist and we can do the syncing
	destinationConfigMap.Data = sourceConfigMap.Data
	if err := r.Update(ctx, destinationConfigMap); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ConfigMapSync{}).
		Owns(&corev1.ConfigMap{}).
		Named("configmapsync").
		Complete(r)
}
