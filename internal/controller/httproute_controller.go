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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/ntentasd/meridian/internal/store"
)

// HTTPRouteReconciler reconciles a HTTPRoute object
type HTTPRouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Store  *store.Store
}

// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes/finalizers,verbs=update

func (r *HTTPRouteReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.V(1).Info("reconciling httproute", "name", req.Name, "namespace", req.Namespace)

	var route gatewayv1.HTTPRoute
	if err := r.Get(ctx, req.NamespacedName, &route); err != nil {
		if apierrors.IsNotFound(err) {
			r.Store.Delete(store.GetKey("HTTPRoute", req.Namespace, req.Name))
			log.V(1).
				Info("deleted httproute from store", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.V(1).Info("upserted httproute to store", "name", req.Name, "namespace", req.Namespace)
	r.Store.Sync(store.GetKey("HTTPRoute", req.Namespace, req.Name), httprouteToEntries(&route))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HTTPRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.HTTPRoute{}).
		Named("httproute").
		Complete(r)
}

func httprouteToEntries(route *gatewayv1.HTTPRoute) []store.RouteEntry {
	var entries []store.RouteEntry
	logoURL := route.Annotations["meridian.ntentas.com/logo-url"]

	for _, host := range route.Spec.Hostnames {
		entries = append(entries, store.RouteEntry{
			UID:       route.UID,
			Name:      route.Name,
			Namespace: route.Namespace,
			Kind:      "HTTPRoute",
			Hostname:  string(host),
			LogoURL:   logoURL,
		})
	}

	return entries
}
