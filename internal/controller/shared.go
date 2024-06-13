package controller

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/biggold1310/silentstorm/api/v1alpha1"
	client2 "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	amc "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	createdBy = "SilentStorm Operator"
	finalizer = "silentstorm.biggold1310.ch/finalizer"
)

// SharedReconciler is able to reconcile Silence and ClusterSilence
type SharedReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	Alertmanager *amc.AlertmanagerAPI
}

func isClean(alertmanagerReferences *[]v1alpha1.AlertmanagerReference) bool {
	clean := true
	for _, alertmanagerRef := range *alertmanagerReferences {
		if alertmanagerRef.SilenceID != "" {
			clean = false
		}
	}
	return clean
}

func (r *SharedReconciler) handleSilence(ctx context.Context, om *metav1.ObjectMeta, alertmanagerSilence v1alpha1.AlertmanagerSilence, alertmanagerReferences *[]v1alpha1.AlertmanagerReference) error {
	var err error
	for _, alertmanagerRef := range *alertmanagerReferences {
		logf := log.FromContext(ctx).WithValues("silenceNamespac", om.GetNamespace(), "silenceName", om.GetName(), "alertmanagerNamespace", alertmanagerRef.Namespace, "alertmanagerName", alertmanagerRef.Name)

		alertmanagerNN := types.NamespacedName{Name: alertmanagerRef.Name, Namespace: alertmanagerRef.Namespace}
		alertmanagerReference := v1alpha1.FindAlertmanagerReference(*alertmanagerReferences, alertmanagerNN)

		alertmanager := &v1alpha1.Alertmanager{}
		err = r.Client.Get(ctx, alertmanagerNN, alertmanager)
		if err != nil {
			if client.IgnoreNotFound(err) == nil {
				logf.Info("Alertmanager not found, cleaning up")
				v1alpha1.RemoveAlertmanagerReference(alertmanagerReferences, alertmanagerNN)
				continue
			}
			return err
		}

		// Configure alertmanager in amc client
		err = r.updateAlertmanagerClient(ctx, *alertmanager)
		if err != nil {
			log.FromContext(ctx).Error(err, "Can't configure alertmanager client")
		}

		var existingSilence *models.GettableSilence
		if alertmanagerReference.SilenceID != "" {
			existingSilence, err = r.getSilenceByID(ctx, alertmanagerReference.SilenceID)
			if err != nil {
				logf.Error(err, "failed to get silence by id", "id", alertmanagerReference.SilenceID)
			}
		}
		if existingSilence == nil {
			existingSilence, err = r.searchSilence(ctx, om.GetUID())
			if err != nil {
				logf.Error(err, "failed to search silences for already existing one")
				continue // Let's skip the rest of the loop as we anyway run into an error
			}
		}

		if om.GetDeletionTimestamp() == nil {
			var silenceID string
			silenceID, err = r.postSilence(ctx, existingSilence, alertmanagerSilence, *om)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed to post silence")
				alertmanagerReference.Status = err.Error()
			} else {
				alertmanagerReference.Status = "Silenced"
				alertmanagerReference.SilenceID = silenceID
			}
		} else {
			err = r.deleteSilence(ctx, existingSilence)
			return err
		}
	}
	return err
}

func (r *SharedReconciler) updateAlertmanagerClient(ctx context.Context, alertmanager v1alpha1.Alertmanager) error {
	aurl, err := url.Parse(alertmanager.Spec.Address)
	if err != nil {
		return err
	}

	schemes := []string{"https"}
	if aurl.Scheme != "" {
		schemes = []string{aurl.Scheme}
	}
	path := "/api/v2"
	if aurl.Path != "" {
		path = aurl.Path
	}

	cr := client2.New(aurl.Host, path, schemes)

	if alertmanager.Spec.Authentication.ServiceAccountName != "" {
		var token string
		token, err = r.getServiceAccountToken(ctx, alertmanager.GetNamespace(), alertmanager.Spec.Authentication.ServiceAccountName)
		if err != nil {
			return err
		}
		client2.BearerToken(token)
	}

	r.Alertmanager.SetTransport(cr)
	return nil
}

// isMatcherEqual compares two slices of Matchers for equality
func isMatcherEqual(m1, m2 models.Matchers) bool {
	if len(m1) != len(m2) {
		return false
	}

	for i := range m1 {
		if !(*m1[i].IsEqual == *m2[i].IsEqual && *m1[i].Name == *m2[i].Name && *m1[i].Value == *m2[i].Value && *m1[i].IsRegex == *m2[i].IsRegex) {
			return false
		}
	}
	return true
}

func (r *SharedReconciler) postSilence(ctx context.Context, existingSilence *models.GettableSilence, alertmanagerSilence v1alpha1.AlertmanagerSilence, om metav1.ObjectMeta) (string, error) {
	ps := &models.PostableSilence{}
	start := strfmt.DateTime(time.Now().UTC().Add(-time.Minute * 5))
	end := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 8))
	comment := fmt.Sprintf("%s    \nSilence UUID: %s", alertmanagerSilence.Comment, string(om.GetUID()))
	creator := createdBy
	ps.Silence.CreatedBy = &creator
	ps.Silence.StartsAt = &start
	ps.Silence.EndsAt = &end
	ps.Silence.Comment = &comment
	ps.Silence.Matchers = alertmanagerSilence.Matchers.ToMatchers()

	if existingSilence != nil {
		if *existingSilence.Comment == comment && *existingSilence.CreatedBy == creator && isMatcherEqual(existingSilence.Matchers, ps.Silence.Matchers) {
			if time.Time(*existingSilence.EndsAt).After(time.Now().UTC().Add(2 * time.Hour)) {
				log.FromContext(ctx).Info("alertmanagerSilence still valid for at least two hours, skipping", "namespace", om.GetNamespace(), "name", om.GetName(), "silenceId", *existingSilence.ID, "expiresAt", *existingSilence.EndsAt)
				return *existingSilence.ID, nil
			}
		}
		ps.ID = *existingSilence.ID
	}

	silenceParams := silence.NewPostSilencesParams().WithContext(ctx).WithSilence(ps)
	postOk, err := r.Alertmanager.Silence.PostSilences(silenceParams)
	if err != nil {
		return "", err
	}
	return postOk.GetPayload().SilenceID, err
}

func (r *SharedReconciler) deleteSilence(ctx context.Context, existingSilence *models.GettableSilence) error {
	if existingSilence != nil {
		silenceParams := silence.NewDeleteSilenceParams().WithContext(ctx).WithSilenceID(strfmt.UUID(*existingSilence.ID))
		deleteOk, err := r.Alertmanager.Silence.DeleteSilence(silenceParams)
		if err != nil {
			return err
		}
		if !deleteOk.IsSuccess() {
			return fmt.Errorf("Failed to delete silence within alertmanager")
		}
		return nil
	}
	return nil
}

func (r *SharedReconciler) getSilenceByID(ctx context.Context, silenceID string) (*models.GettableSilence, error) {
	getSilenceParams := silence.NewGetSilenceParams().WithContext(ctx).WithSilenceID(strfmt.UUID(silenceID))
	getOk, err := r.Alertmanager.Silence.GetSilence(getSilenceParams)
	if err != nil {
		return nil, err
	}
	return getOk.GetPayload(), nil
}

func (r *SharedReconciler) searchSilence(ctx context.Context, uid types.UID) (*models.GettableSilence, error) {
	getSilenceParams := silence.NewGetSilencesParams().WithContext(ctx)
	getOk, err := r.Alertmanager.Silence.GetSilences(getSilenceParams)
	if err != nil {
		return nil, err
	}
	if getOk.Payload == nil {
		return nil, nil
	}

	existingSilences := []models.GettableSilence{}
	for _, silence := range getOk.Payload {
		// Skip expired silences
		if time.Time(*silence.EndsAt).Before(time.Now()) || *silence.Status.State == models.SilenceStatusStateExpired {
			continue
		}
		if *silence.CreatedBy != "SilentStorm Operator" {
			continue
		}
		if !strings.Contains(*silence.Comment, string(uid)) {
			continue
		}

		existingSilences = append(existingSilences, *silence)
	}

	if len(existingSilences) == 1 {
		return &existingSilences[0], nil
	} else if len(existingSilences) == 0 {
		return nil, nil
	}
	return nil, errors.New("more than one silence found")
}

func (r *SharedReconciler) getServiceAccountToken(ctx context.Context, namespace, saName string) (string, error) {
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
