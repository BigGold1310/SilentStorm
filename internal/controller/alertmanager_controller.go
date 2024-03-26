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
	"fmt"
	"net/url"
	"time"

	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	"github.com/go-openapi/strfmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clientruntime "github.com/go-openapi/runtime/client"

	alertmanagerclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
)

// AlertmanagerReconciler reconciles a Alertmanager object
type AlertmanagerReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	Alertmanager *alertmanagerclient.AlertmanagerAPI
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

	err = r.initAlertmanagerClient(ctx, *alertmanager)
	if err != nil {
		return ctrl.Result{}, err
	}

	silenceList := silentstormv1alpha1.SilenceList{}
	err = r.Client.List(ctx, &silenceList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return ctrl.Result{}, err
	}
	clusterSilenceList := silentstormv1alpha1.ClusterSilenceList{}
	err = r.Client.List(ctx, &clusterSilenceList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, clusterSilence := range clusterSilenceList.Items {
		oldStatus := clusterSilence.Status.DeepCopy()

		alertmanagerReference := silentstormv1alpha1.FindAlertmanagerReference(clusterSilence.Status.AlertmanagerReferences, alertmanager.UID)
		if alertmanagerReference == nil {
			if silentstormv1alpha1.SetAlertmanagerReference(&clusterSilence.Status.AlertmanagerReferences, silentstormv1alpha1.NewAlertmanagerReference(*alertmanager)) {
				alertmanagerReference = silentstormv1alpha1.FindAlertmanagerReference(clusterSilence.Status.AlertmanagerReferences, alertmanager.UID)
			}
		}

		r.Alertmanager.Silence.GetSilences()
		silence.NewGetSilencesParams().SetFilter()

		start := strfmt.DateTime(time.Now().Add(-time.Minute * 5))
		end := strfmt.DateTime(time.Now().Add(time.Hour * 2))
		creator := "SilentStorm Operator"
		ps := &models.PostableSilence{
			Silence: models.Silence{
				Matchers:  clusterSilence.Spec.Matchers.ToMatchers(),
				StartsAt:  &start,
				EndsAt:    &end,
				CreatedBy: &creator,
				Comment:   &clusterSilence.Spec.Comment,
			},
		}
		if alertmanagerReference.SilenceID != "" {
			ps.ID = alertmanagerReference.SilenceID
		}
		silenceParams := silence.NewPostSilencesParams().WithContext(ctx).WithSilence(ps)
		postOk, err := r.Alertmanager.Silence.PostSilences(silenceParams)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to post silence", "namespace", clusterSilence.GetNamespace(), "name", clusterSilence.GetName())
			alertmanagerReference.Status = err.Error()
		} else {
			alertmanagerReference.Status = "Silenced" // TODO: introduce consts for better status tracking/consistency
			alertmanagerReference.SilenceID = postOk.GetPayload().SilenceID
		}

		if !equality.Semantic.DeepEqual(oldStatus, &clusterSilence.Status) {
			err = r.Client.Status().Update(ctx, &clusterSilence)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed update silence status with id", "namespace", clusterSilence.GetNamespace(), "name", clusterSilence.GetName(), "silenceId", postOk.GetPayload().SilenceID)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *AlertmanagerReconciler) initAlertmanagerClient(ctx context.Context, alertmanager silentstormv1alpha1.Alertmanager) error {
	if r.Alertmanager == nil {
		// Lets initialize the Alertmanager client. This functionality is mainly to allow unit testing.
		url, err := url.Parse(alertmanager.Spec.Address)
		if err != nil {
			panic(err)
		}

		schemes := []string{"https"}
		if url.Scheme != "" {
			schemes = []string{url.Scheme}
		}
		path := "/api/v2"
		if url.Path != "" {
			path = url.Path
		}

		cr := clientruntime.New(url.Host, path, schemes)

		if alertmanager.Spec.Authentication.ServiceAccount != "" {
			token, err := r.getServiceAccountToken(ctx, alertmanager.GetNamespace(), alertmanager.Spec.Authentication.ServiceAccount)
			if err != nil {
				return err
			}
			clientruntime.BearerToken(token)
		}

		c := alertmanagerclient.New(cr, strfmt.Default)
		r.Alertmanager = c
	}
	return nil
}

func (r *AlertmanagerReconciler) getServiceAccountToken(ctx context.Context, namespace, saName string) (string, error) {
	// Fetch the Secret associated with the ServiceAccount
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: fmt.Sprintf("%s-token", saName), Namespace: namespace}, secret)
	if err != nil {
		return "", err
	}

	// Extract and return the token from the Secret
	token, found := secret.Data[corev1.ServiceAccountTokenKey]
	if !found {
		return "", fmt.Errorf("service account token not found")
	}
	return string(token), nil
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
