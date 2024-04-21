package utils

import (
	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewTestClient(initObjs ...client.Object) (client.Client, *runtime.Scheme) {
	scheme := runtime.NewScheme()
	silentstormv1alpha1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).WithStatusSubresource(initObjs...).Build(), scheme
}
