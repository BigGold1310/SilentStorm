package testdata

import (
	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateAlertmanager(name string) *silentstormv1alpha1.Alertmanager {
	return &silentstormv1alpha1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
		},
		Spec: silentstormv1alpha1.AlertmanagerSpec{},
	}
}
