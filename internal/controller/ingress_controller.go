/*
Copyright 2026.

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

	"github.com/ntentasd/meridian/internal/store"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Store  *store.Store
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.V(1).Info("reconciling ingress", "name", req.Name, "namespace", req.Namespace)

	var ing networkingv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ing); err != nil {
		if apierrors.IsNotFound(err) {
			r.Store.Delete(store.GetKey("Ingress", req.Namespace, req.Name))
			log.V(1).
				Info("deleted ingress from store", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	log.V(1).Info("upserted ingress to store", "name", req.Name, "namespace", req.Namespace)
	r.Store.Sync(store.GetKey("Ingress", req.Namespace, req.Name), ingressToEntries(&ing))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkingv1.Ingress{}).
		Named("ingress").
		Complete(r)
}

// TODO: add enricher
func ingressToEntries(ing *networkingv1.Ingress) []store.RouteEntry {
	var entries []store.RouteEntry
	logoURL := ing.Annotations["meridian.ntentas.com/logo-url"]
	// Add other annotations as needed (owner, desc, etc.)

	for _, rule := range ing.Spec.Rules {
		if rule.Host == "" {
			continue
		}
		entries = append(entries, store.RouteEntry{
			UID:       ing.UID,
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Kind:      "Ingress",
			Hostname:  rule.Host,
			LogoURL:   logoURL,
		})
	}

	return entries
}
