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
	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
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
			mockCtrl       *gomock.Controller
			amcmock        *amc.AlertmanagerAPI
			mockSilence    *mock_silence.MockClientService
			mockAlert      *mock_alert.MockClientService
			mockGeneral    *mock_general.MockClientService
			mockReceiver   *mock_receiver.MockClientService
			mockAlertgroup *mock_alertgroup.MockClientService
			mockTransport  *mock_runtime.MockClientTransport
			ctx            context.Context
		)
		BeforeEach(func() {
			ctx = context.Background()
			mockCtrl = gomock.NewController(GinkgoT())
			mockSilence = mock_silence.NewMockClientService(mockCtrl)
			mockAlert = mock_alert.NewMockClientService(mockCtrl)
			mockGeneral = mock_general.NewMockClientService(mockCtrl)
			mockReceiver = mock_receiver.NewMockClientService(mockCtrl)
			mockAlertgroup = mock_alertgroup.NewMockClientService(mockCtrl)
			mockTransport = mock_runtime.NewMockClientTransport(mockCtrl)
			amcmock = &amc.AlertmanagerAPI{
				Alert:      mockAlert,
				Alertgroup: mockAlertgroup,
				Silence:    mockSilence,
				General:    mockGeneral,
				Receiver:   mockReceiver,
				Transport:  mockTransport,
			}
			mockSilence.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			mockAlertgroup.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			mockSilence.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			mockGeneral.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			mockReceiver.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
		})

		It("should successfully reconcile the resource", func() {
			mockAlert.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()

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
		It("should register itself on the ClusterSilence", func() {
			mockAlert.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			clusterSilence := testdata.GenerateClusterSilence("clustersilence-1")
			clusterSilence.ObjectMeta.SetLabels(map[string]string{"silence": "test-silence"})

			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(clusterSilence, alertmanager1)
			reconciler := &controller.AlertmanagerReconciler{SharedReconciler: controller.SharedReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
			updatedClusterSilence := silentstormv1alpha1.ClusterSilence{}
			err = client.Get(ctx, types.NamespacedName{Name: "clustersilence-1"}, &updatedClusterSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updatedClusterSilence.Status.AlertmanagerReferences)).Should(Equal(1))
			Expect(updatedClusterSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
		})
		It("multiple Alertmanager should register on ClusterSilence", func() {
			mockAlert.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			clusterSilence := testdata.GenerateClusterSilence("clustersilence-1")
			clusterSilence.ObjectMeta.SetLabels(map[string]string{"silence": "test-silence"})

			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}
			alertmanager2 := testdata.GenerateAlertmanager("alertmanager-2")
			alertmanager2.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(clusterSilence, alertmanager1, alertmanager2)
			reconciler := &controller.AlertmanagerReconciler{SharedReconciler: controller.SharedReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})
			Expect(err).NotTo(HaveOccurred())

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-2", Namespace: testdata.Namespace}})
			Expect(err).NotTo(HaveOccurred())

			updatedClusterSilence := silentstormv1alpha1.ClusterSilence{}
			err = client.Get(ctx, types.NamespacedName{Name: "clustersilence-1"}, &updatedClusterSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updatedClusterSilence.Status.AlertmanagerReferences)).Should(Equal(2))
			Expect(updatedClusterSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
			Expect(updatedClusterSilence.Status.AlertmanagerReferences[1].Name).Should(Equal("alertmanager-2"))
		})
		It("should register itself on the Silence", func() {
			mockAlert.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			silence := testdata.GenerateSilence("silence-1")
			silence.ObjectMeta.SetLabels(map[string]string{"silence": "test-silence"})

			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(silence, alertmanager1)
			reconciler := &controller.AlertmanagerReconciler{SharedReconciler: controller.SharedReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
			updatedSilence := silentstormv1alpha1.Silence{}
			err = client.Get(ctx, types.NamespacedName{Name: "silence-1", Namespace: testdata.Namespace}, &updatedSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updatedSilence.Status.AlertmanagerReferences)).Should(Equal(1))
			Expect(updatedSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
		})
		It("multiple Alertmanager should register on Silence", func() {
			mockAlert.EXPECT().SetTransport(gomock.Any()).Return().AnyTimes()
			silence := testdata.GenerateSilence("silence-1")
			silence.ObjectMeta.SetLabels(map[string]string{"silence": "test-silence"})

			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}
			alertmanager2 := testdata.GenerateAlertmanager("alertmanager-2")
			alertmanager2.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(silence, alertmanager1, alertmanager2)
			reconciler := &controller.AlertmanagerReconciler{SharedReconciler: controller.SharedReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})
			Expect(err).NotTo(HaveOccurred())

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-2", Namespace: testdata.Namespace}})
			Expect(err).NotTo(HaveOccurred())

			updateSilence := silentstormv1alpha1.Silence{}
			err = client.Get(ctx, types.NamespacedName{Name: "silence-1", Namespace: testdata.Namespace}, &updateSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updateSilence.Status.AlertmanagerReferences)).Should(Equal(2))
			Expect(updateSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
			Expect(updateSilence.Status.AlertmanagerReferences[1].Name).Should(Equal("alertmanager-2"))
		})
	})
})
