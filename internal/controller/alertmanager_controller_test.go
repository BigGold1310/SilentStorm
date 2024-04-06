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
	mock_alert "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alert"
	mock_alertgroup "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/alertgroup"
	mock_general "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/general"
	mock_receiver "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/receiver"
	mock_runtime "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/runtime"
	mock_silence "github.com/biggold1310/silentstorm/internal/mocks/alertmanager/silence"
	. "github.com/onsi/ginkgo/v2"

	amc "github.com/prometheus/alertmanager/api/v2/client"
	gm "go.uber.org/mock/gomock"
)

var _ = Describe("Alertmanager Controller", func() {
	Context("When reconciling a resource", func() {
		var (
			mockCtrl *gm.Controller // gomock struct
			amcmock  *amc.AlertmanagerAPI
		)
		BeforeEach(func() {
			mockCtrl = gm.NewController(GinkgoT())
			amcmock = &amc.AlertmanagerAPI{
				Alert:      mock_alert.NewMockClientService(mockCtrl),
				Alertgroup: mock_alertgroup.NewMockClientService(mockCtrl),
				Silence:    mock_silence.NewMockClientService(mockCtrl),
				General:    mock_general.NewMockClientService(mockCtrl),
				Receiver:   mock_receiver.NewMockClientService(mockCtrl),
				Transport:  mock_runtime.NewMockClientTransport(mockCtrl),
			}
		})

		It("should successfully reconcile the resource", func() {
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
			reconciler := &AlertmanagerReconciler{Client:, Alertmanager: amcmock, Scheme: }
		})
	})
})
