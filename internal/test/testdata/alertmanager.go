package testdata

import (
	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateAlertmanager(name string) *silentstormv1alpha1.Alertmanager {
	return &silentstormv1alpha1.Alertmanager{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "silentstorm.biggold1310.ch/v1alpha1",
			Kind:       "Alertmanager",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
		},
		Spec: silentstormv1alpha1.AlertmanagerSpec{},
	}
}

// GenerateClusterSilence generates a ClusterSilence object for testing purposes.
func GenerateClusterSilence(name string) *silentstormv1alpha1.ClusterSilence {
	return &silentstormv1alpha1.ClusterSilence{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "silentstorm.biggold1310.ch/v1alpha1",
			Kind:       "ClusterSilence",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
		},
		Spec: silentstormv1alpha1.ClusterSilenceSpec{
			AlertmanagerSilence: silentstormv1alpha1.AlertmanagerSilence{
				Matchers: silentstormv1alpha1.Matchers{},
				Creator:  "test",
				Comment:  "test",
			},
		},
	}
}

// GenerateSilence generates a Silence object for testing purposes.
func GenerateSilence(name string) *silentstormv1alpha1.Silence {
	return &silentstormv1alpha1.Silence{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "silentstorm.biggold1310.ch/v1alpha1",
			Kind:       "Silence",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
		},
		Spec: silentstormv1alpha1.SilenceSpec{
			AlertmanagerSilence: silentstormv1alpha1.AlertmanagerSilence{
				Matchers: silentstormv1alpha1.Matchers{},
				Creator:  "test",
				Comment:  "test",
			},
		},
	}
}
