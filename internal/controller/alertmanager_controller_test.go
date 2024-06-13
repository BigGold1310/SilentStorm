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

package controller_test

import (
	"context"

	"github.com/biggold1310/silentstorm/internal/controller"

	mock_alert "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alert"
	mock_alertgroup "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alertgroup"
	mock_general "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/general"
	mock_receiver "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/receiver"
	mock_runtime "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/runtime"
	mock_silence "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/silence"
	"github.com/biggold1310/silentstorm/internal/test/testdata"
	testutils "github.com/biggold1310/silentstorm/internal/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amc "github.com/prometheus/alertmanager/api/v2/client"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Alertmanager Controller", func() {
	Context("When reconciling a resource", func() {
		var (
			mockCtrl    *gomock.Controller
			amcmock     *amc.AlertmanagerAPI
			mockSilence *mock_silence.MockClientService
			ctx         context.Context
		)
		BeforeEach(func() {
			ctx = context.Background()
			mockCtrl = gomock.NewController(GinkgoT())
			mockSilence = mock_silence.NewMockClientService(mockCtrl)
			amcmock = &amc.AlertmanagerAPI{
				Alert:      mock_alert.NewMockClientService(mockCtrl),
				Alertgroup: mock_alertgroup.NewMockClientService(mockCtrl),
				Silence:    mockSilence,
				General:    mock_general.NewMockClientService(mockCtrl),
				Receiver:   mock_receiver.NewMockClientService(mockCtrl),
				Transport:  mock_runtime.NewMockClientTransport(mockCtrl),
			}
		})

		It("should successfully reconcile the resource", func() {
			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(alertmanager1)
			reconciler := &controller.AlertmanagerReconciler{SharedReconciler: controller.SharedReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
