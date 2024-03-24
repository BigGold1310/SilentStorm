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

package v1alpha1

import (
	"github.com/prometheus/alertmanager/api/v2/models"
	"k8s.io/apimachinery/pkg/types"
)

// AlertmanagerSilence holds the common declaration required in ClusterSilence and Silence specs.
type AlertmanagerSilence struct {
	Matchers `json:"matchers"`
	Creator  string `json:"creator,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type Matcher struct {
	// isEqual allows to switch the matcher between equality and non-equality. Default is equality (true)
	// +kubebuilder:default=true
	IsEqual bool `json:"isEqual,omitempty"`
	// isRegex allows to switch the matcher between regex and non-regex. Default is false.
	// +kubebuilder:default=false
	IsRegex bool `json:"isRegex"`
	// name, within alertmanager usually known as label name/key.
	Name string `json:"name"`
	// value, within alertmanager usually known as label value/value.
	Value string `json:"value"`
}

// ToMatcher converts the internal AlertmanagerSilence matcher to the openapi matcher.
func (m Matcher) ToMatcher() *models.Matcher {
	return &models.Matcher{
		Name:    &m.Name,
		Value:   &m.Value,
		IsEqual: &m.IsEqual,
		IsRegex: &m.IsRegex,
	}
}

type Matchers []Matcher

// ToMatchers converts the internal AlertmanagerSilence matchers to the openapi matchers.
func (m Matchers) ToMatchers() []*models.Matcher {
	matchers := make([]*models.Matcher, len(m))
	for i, matcher := range m {
		matchers[i] = matcher.ToMatcher()
	}
	return matchers
}

// AlertmanagerReference tracks the Silence status within an alertmanager instance.
type AlertmanagerReference struct {
	// UID of the alertmanager instance which aknowledged the silence.
	UID types.UID `json:"uid,omitempty"`
	// SilenceID within the aknowledged alertmanager instance.
	SilenceID string `json:"silenceId,omitempty"`
	// Name of the alertmanager instance which aknowledged the silence.
	Name string `json:"name,omitempty"`
	// Namespace of the alertmanager instance which aknowledged the silence.
	Namespace string `json:"namespace,omitempty"`
	// Status of the silence within the aknowledged alertmanager instance.
	Status string `json:"status,omitempty"`
}

// FindAlertmanagerReference returns the AlertmanagerReference for the given UID.
func FindAlertmanagerReference(ars []AlertmanagerReference, uid types.UID) *AlertmanagerReference {
	for i := range ars {
		if ars[i].UID == uid {
			return &ars[i]
		}
	}
	return nil
}

// RemoveAlertmanagerReference removes the corresponding conditionType from conditions if present. Returns
// true if it was present and got removed.
func RemoveAlertmanagerReference(ars *[]AlertmanagerReference, uid types.UID) (removed bool) {
	if ars == nil || len(*ars) == 0 {
		return false
	}
	newARs := make([]AlertmanagerReference, 0, len(*ars)-1)
	for _, ar := range *ars {
		if ar.UID != uid {
			newARs = append(newARs, ar)
		}
	}
	removed = len(newARs) != len(*ars)
	*ars = newARs
	return removed
}

// SetAlertmanagerReference sets (adds or updates) the corresponding AlertmanagerReference in AlertmanagerReferences returns true if AlertmanagerReferences was changed
// conditions must be non-nil.
//  1. if the AlertmanagerReference of the specified UID already exists (all fields of the existing AlertmanagerReference are updated)
//  2. if a AlertmanagerReference of the specified type does not exist it gets appended to AlertmanagerReferences
//
// Note: The Name, Namespace and UID fields of the AlertmanagerReference are immutable.
func SetAlertmanagerReference(ars *[]AlertmanagerReference, newAR AlertmanagerReference) (changed bool) {
	if ars == nil {
		return false
	}
	existingAR := FindAlertmanagerReference(*ars, newAR.UID)
	if existingAR == nil {
		*ars = append(*ars, newAR)
		return true
	}

	if existingAR.SilenceID != newAR.SilenceID {
		existingAR.SilenceID = newAR.SilenceID
		changed = true
	}
	if existingAR.Status != newAR.Status {
		existingAR.Status = newAR.Status
		changed = true
	}
	return changed
}

func NewAlertmanagerReference(am Alertmanager) AlertmanagerReference {
	return AlertmanagerReference{
		UID:       am.GetUID(),
		Name:      am.GetName(),
		Namespace: am.GetNamespace(),
	}
}
