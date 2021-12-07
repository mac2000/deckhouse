/*
Copyright 2021 Flant JSC

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

package hooks

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = FDescribe("Prometheus hooks :: grafana notification channels ::", func() {
	f := HookExecutionConfigInit(`
{
  "global": {
    "enabledModules": [],
    "discovery":{
		"clusterDomain": "cluster.my"
    }
  },
  "prometheus":{
    "internal":{
      "grafana":{}
    }
 }
}`, ``)
	f.RegisterCRD("deckhouse.io", "v1alpha1", "GrafanaAlertsChannel", false)

	getChannelsFromValues := func(*HookExecutionConfig) []GrafanaAlertsChannel {
		channels := f.ValuesGet("prometheus.internal.grafana.alertsChannels").Array()

		res := make([]GrafanaAlertsChannel, 0)

		for _, raw := range channels {
			c := GrafanaAlertsChannel{}
			err := json.Unmarshal([]byte(raw.Raw), &c)
			Expect(err).ToNot(HaveOccurred())
			res = append(res, c)
		}

		return res
	}

	Context("Empty cluster", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Does not set any channels in values", func() {
			Expect(f).To(ExecuteSuccessfully())

			Expect(getChannelsFromValues(f)).To(HaveLen(0))
		})

		Context("Add channel", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(`
---
apiVersion: deckhouse.io/v1alpha1
kind: GrafanaAlertsChannel
metadata:
  name: test
spec:
  type: prometheus-alertmanager
  alertManager:
    address: "http://some-alert-manager"
    auth:
      basic:
        username: user
        password: password
`, 0))
				f.RunHook()
			})

			It("Should store channel in values", func() {
				Expect(f).To(ExecuteSuccessfully())
				channels := getChannelsFromValues(f)

				Expect(channels).To(HaveLen(1))

				Expect(channels[0]).To(Equal(GrafanaAlertsChannel{
					OrgID:                 1,
					Type:                  alertManagerGrafanaAlertChannelType,
					Name:                  "test",
					UID:                   "test",
					IsDefault:             false,
					DisableResolveMessage: false,
					SendReminder:          false,
					Frequency:             time.Duration(0),
					Settings: map[string]interface{}{
						"url":           "http://some-alert-manager",
						"basicAuthUser": "user",
					},
					SecureSettings: map[string]interface{}{
						"basicAuthPassword": "password",
					},
				}))
			})

			Context("Deleting channel", func() {
				BeforeEach(func() {
					f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(``, 1))
					f.RunHook()
				})

				It("Should delete GrafanaAdditionalDatasource from values", func() {
					Expect(f).To(ExecuteSuccessfully())
					Expect(getChannelsFromValues(f)).To(HaveLen(0))
				})
			})

			Context("Updating channel", func() {
				BeforeEach(func() {
					f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(`
---
apiVersion: deckhouse.io/v1alpha1
kind: GrafanaAlertsChannel
metadata:
  name: test
spec:
  type: prometheus-alertmanager
  alertManager:
    address: "https://another-url"
    auth:
      basic:
        username: user
        password: another-password
`, 1))
					f.RunHook()
				})

				It("Should update GrafanaAdditionalDatasource in values", func() {
					Expect(f).To(ExecuteSuccessfully())
					channels := getChannelsFromValues(f)

					Expect(channels[0]).To(Equal(Expect(channels[0]).To(Equal(GrafanaAlertsChannel{
						OrgID:                 1,
						Type:                  alertManagerGrafanaAlertChannelType,
						Name:                  "test",
						UID:                   "test",
						IsDefault:             false,
						DisableResolveMessage: false,
						SendReminder:          false,
						Frequency:             time.Duration(0),
						Settings: map[string]interface{}{
							"url":           "https://another-url",
							"basicAuthUser": "user",
						},
						SecureSettings: map[string]interface{}{
							"basicAuthPassword": "another-password",
						},
					}))))
				})
			})
		})

		Context("Enable flant integration module", func() {
			BeforeEach(func() {
				f.ValuesSetFromYaml("global.enabledModules", []byte(`["flant-integration"]`))
				f.RunHook()
			})

			It("Should store madison alerts channel in values with url only", func() {
				Expect(f).To(ExecuteSuccessfully())
				channels := getChannelsFromValues(f)

				Expect(channels).To(HaveLen(1))

				Expect(channels[0]).To(Equal(GrafanaAlertsChannel{
					OrgID:                 1,
					Type:                  alertManagerGrafanaAlertChannelType,
					Name:                  madisonAlertChannelName,
					UID:                   madisonAlertChannelName,
					IsDefault:             false,
					DisableResolveMessage: false,
					SendReminder:          false,
					Frequency:             time.Duration(0),
					Settings: map[string]interface{}{
						"url": "http://madison-proxy.d8-monitoring.svc.cluster.my",
					},
					SecureSettings: make(map[string]interface{}),
				}))
			})
		})
	})

	//	Context("Cluster with GrafanaAdditionalDatasource", func() {
	//		BeforeEach(func() {
	//			f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(`
	//---
	//apiVersion: deckhouse.io/v1
	//kind: GrafanaAdditionalDatasource
	//metadata:
	//  name: test
	//spec:
	//  url: /abc
	//  type: test
	//  access: Proxy
	//---
	//apiVersion: deckhouse.io/v1
	//kind: GrafanaAdditionalDatasource
	//metadata:
	//  name: test-next
	//spec:
	//  url: /def
	//  type: test-next
	//  access: Direct
	//`, 2))
	//			f.RunHook()
	//		})
	//
	//		It("Should synchronize the GrafanaAdditionalDatasource to values", func() {
	//			Expect(f).To(ExecuteSuccessfully())
	//			Expect(f.ValuesGet("prometheus.internal.grafana.additionalDatasources").String()).To(MatchJSON(`
	//[{
	//   "access": "proxy",
	//   "editable": false,
	//   "isDefault": false,
	//   "name": "test",
	//   "orgId": 1,
	//   "type": "test",
	//   "url": "/abc",
	//   "uuid": "test",
	//   "version": 1
	//},{
	//   "access": "direct",
	//   "editable": false,
	//   "isDefault": false,
	//   "name": "test-next",
	//   "orgId": 1,
	//   "type": "test-next",
	//   "url": "/def",
	//   "uuid": "test-next",
	//   "version": 1
	//}]`))
	//		})
	//	})
})
