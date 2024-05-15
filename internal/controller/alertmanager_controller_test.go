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
	silentstormv1alpha1 "github.com/biggold1310/silentstorm/api/v1alpha1"
	mock_alert "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alert"
	mock_alertgroup "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alertgroup"
	mock_general "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/general"
	mock_receiver "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/receiver"
	mock_runtime "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/runtime"
	mock_silence "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/silence"
	"github.com/biggold1310/silentstorm/internal/test/testdata"
	testutils "github.com/biggold1310/silentstorm/internal/test/utils"
	"github.com/go-openapi/strfmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amc "github.com/prometheus/alertmanager/api/v2/client"
	amcsilence "github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
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
			reconciler := &AlertmanagerReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should process a new ClusterSilence successfully", func() {
			silenceID := string(uuid.NewUUID())
			mockSilence.EXPECT().GetSilences(gomock.Any()).Return(&amcsilence.GetSilencesOK{}, nil).AnyTimes()
			mockSilence.EXPECT().PostSilences(gomock.Any()).Return(&amcsilence.PostSilencesOK{Payload: &amcsilence.PostSilencesOKBody{SilenceID: silenceID}}, nil).AnyTimes()

			clusterSilence := &silentstormv1alpha1.ClusterSilence{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-silence",
					UID:    uuid.NewUUID(),
					Labels: map[string]string{"silence": "test-silence"},
				},
				Spec: silentstormv1alpha1.ClusterSilenceSpec{
					AlertmanagerSilence: silentstormv1alpha1.AlertmanagerSilence{
						Matchers: silentstormv1alpha1.Matchers{
							{
								Name:  "test-matcher",
								Value: "test-value",
							},
						},
						Creator: "user1",
						Comment: "test comment",
					},
				},
			}
			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(clusterSilence, alertmanager1)
			reconciler := &AlertmanagerReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
			updatedClusteerSilence := silentstormv1alpha1.ClusterSilence{}
			err = client.Get(ctx, types.NamespacedName{Namespace: "", Name: "test-silence"}, &updatedClusteerSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updatedClusteerSilence.Status.AlertmanagerReferences)).Should(Equal(1))
			Expect(updatedClusteerSilence.Status.AlertmanagerReferences[0].SilenceID).Should(Equal(silenceID))
			Expect(updatedClusteerSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
		})

		It("should not create a new silence if there is already an existing one", func() {
			silenceID := uuid.NewUUID()

			mockSilence.EXPECT().GetSilences(gomock.Any()).Return(&amcsilence.GetSilencesOK{Payload: models.GettableSilences{&models.GettableSilence{
				ID:        testutils.ToPtr(string(silenceID)),
				Status:    &models.SilenceStatus{State: testutils.ToPtr(models.SilenceStatusStateActive)},
				UpdatedAt: nil,
				Silence: models.Silence{
					Comment:   testutils.ToPtr(fmt.Sprintf("test comment    \nSilence UUID: %s", silenceID)),
					CreatedBy: testutils.ToPtr("SilentStorm Operator"),
					EndsAt:    testutils.ToPtr(strfmt.DateTime(time.Now().Add(time.Hour * 3))),
					Matchers: models.Matchers{
						&models.Matcher{
							IsEqual: testutils.ToPtr(false),
							IsRegex: testutils.ToPtr(false),
							Name:    testutils.ToPtr("test-matcher"),
							Value:   testutils.ToPtr("test-value"),
						},
					},
					StartsAt: testutils.ToPtr(strfmt.DateTime(time.Now().Add(-time.Hour * 2))),
				},
			}}}, nil).Times(1)

			clusterSilence := &silentstormv1alpha1.ClusterSilence{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-silence",
					UID:    silenceID,
					Labels: map[string]string{"silence": "test-silence"},
				},
				Spec: silentstormv1alpha1.ClusterSilenceSpec{
					AlertmanagerSilence: silentstormv1alpha1.AlertmanagerSilence{
						Matchers: silentstormv1alpha1.Matchers{
							{
								Name:  "test-matcher",
								Value: "test-value",
							},
						},
						Creator: "user1",
						Comment: "test comment",
					},
				},
			}
			alertmanager1 := testdata.GenerateAlertmanager("alertmanager-1")
			alertmanager1.Spec.SilenceSelector = metav1.LabelSelector{
				MatchLabels:      map[string]string{"silence": "test-silence"},
				MatchExpressions: nil,
			}

			client, scheme := testutils.NewTestClient(clusterSilence, alertmanager1)
			reconciler := &AlertmanagerReconciler{Client: client, Alertmanager: amcmock, Scheme: scheme}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "alertmanager-1", Namespace: testdata.Namespace}})

			Expect(err).NotTo(HaveOccurred())
			updatedClusteerSilence := silentstormv1alpha1.ClusterSilence{}
			err = client.Get(ctx, types.NamespacedName{Namespace: "", Name: "test-silence"}, &updatedClusteerSilence)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(updatedClusteerSilence.Status.AlertmanagerReferences)).Should(Equal(1))
			Expect(updatedClusteerSilence.Status.AlertmanagerReferences[0].SilenceID).Should(Equal(string(silenceID)))
			Expect(updatedClusteerSilence.Status.AlertmanagerReferences[0].Name).Should(Equal("alertmanager-1"))
		})
	})
})
