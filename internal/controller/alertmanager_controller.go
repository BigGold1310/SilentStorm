/*
Copyright 2024.

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
	"time"

	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SilenceReconciler reconciles a Alertmanager object
type AlertmanagerReconciler struct {
	SharedReconciler
}

//+kubebuilder:rbac:groups=silentstorm.biggold1310.ch,resources=alertmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=silentstorm.biggold1310.ch,resources=alertmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=silentstorm.biggold1310.ch,resources=alertmanagers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AlertmanagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	_ = log.FromContext(ctx)

	alertmanager := &silentstormv1alpha1.Alertmanager{}
	err = r.Client.Get(ctx, req.NamespacedName, alertmanager)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(&alertmanager.Spec.SilenceSelector)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to parse silence selector", "namespace", alertmanager.GetNamespace(), "name", alertmanager.GetName())
	}

	// Configure alertmanager in amc client
	err = r.updateAlertmanagerClient(ctx, *alertmanager)
	if err != nil {
		log.FromContext(ctx).Error(err, "Can't configure alertmanager client")
	}

	// TODO: Check alertmanager status (if we can reach it) and provide info to user
	// r.Alertmanager.General.GetStatus()

	silenceList := silentstormv1alpha1.SilenceList{}
	err = r.Client.List(ctx, &silenceList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, silence := range silenceList.Items {
		oldStatus := silence.Status.DeepCopy()

		silentstormv1alpha1.SetAlertmanagerReference(&silence.Status.AlertmanagerReferences, silentstormv1alpha1.NewAlertmanagerReference(*alertmanager))

		if !equality.Semantic.DeepEqual(oldStatus, &silence.Status) {
			err = r.Client.Status().Update(ctx, &silence)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed to update silence with alertmanager reference", "namespace", silence.GetNamespace(), "name", silence.GetName())
			}
		}
	}

	clusterSilenceList := silentstormv1alpha1.ClusterSilenceList{}
	err = r.Client.List(ctx, &clusterSilenceList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, clusterSilence := range clusterSilenceList.Items {
		oldStatus := clusterSilence.Status.DeepCopy()

		silentstormv1alpha1.SetAlertmanagerReference(&clusterSilence.Status.AlertmanagerReferences, silentstormv1alpha1.NewAlertmanagerReference(*alertmanager))

		if !equality.Semantic.DeepEqual(oldStatus, &clusterSilence.Status) {
			err = r.Client.Status().Update(ctx, &clusterSilence)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed to update silence with alertmanager reference", "namespace", clusterSilence.GetNamespace(), "name", clusterSilence.GetName())
			}
		}
	}

	return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertmanagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&silentstormv1alpha1.Alertmanager{}).
		Watches(&silentstormv1alpha1.ClusterSilence{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var reconcileRequests []reconcile.Request
			alertmanagerList := &silentstormv1alpha1.AlertmanagerList{}
			client := mgr.GetClient()

			err := client.List(ctx, alertmanagerList)
			if err != nil {
				return reconcileRequests
			}

			if silence, ok := obj.(*silentstormv1alpha1.ClusterSilence); ok {
				for _, alertmanager := range alertmanagerList.Items {
					selector, err := metav1.LabelSelectorAsSelector(&alertmanager.Spec.SilenceSelector)
					if err != nil {
						log.FromContext(ctx).Error(err, "failed to parse silence selector", "namespace", alertmanager.GetNamespace(), "name", alertmanager.GetName())
					}
					if selector.Matches(labels.Set(silence.GetLabels())) {
						rec := reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      alertmanager.GetName(),
								Namespace: alertmanager.GetNamespace(),
							},
						}
						reconcileRequests = append(reconcileRequests, rec)
					}
				}
			}
			return reconcileRequests
		})).
		Watches(&silentstormv1alpha1.Silence{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var reconcileRequests []reconcile.Request
			alertmanagerList := &silentstormv1alpha1.AlertmanagerList{}
			client := mgr.GetClient()

			err := client.List(ctx, alertmanagerList)
			if err != nil {
				return reconcileRequests
			}

			if silence, ok := obj.(*silentstormv1alpha1.Silence); ok {
				for _, alertmanager := range alertmanagerList.Items {
					selector, err := metav1.LabelSelectorAsSelector(&alertmanager.Spec.SilenceSelector)
					if err != nil {
						log.FromContext(ctx).Error(err, "failed to parse silence selector", "namespace", alertmanager.GetNamespace(), "name", alertmanager.GetName())
					}
					if selector.Matches(labels.Set(silence.GetLabels())) {
						rec := reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      alertmanager.GetName(),
								Namespace: alertmanager.GetNamespace(),
							},
						}
						reconcileRequests = append(reconcileRequests, rec)
					}
				}
			}
			return reconcileRequests
		})).
		Complete(r)
}
