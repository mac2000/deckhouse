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

package template_tests

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/helm"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "")
}

const (
	globalValues = `
  enabledModules: ["vertical-pod-autoscaler-crd"]
  modules:
    placement: {}
  modulesImages:
    registry: registry.deckhouse.io
    registryDockercfg: cfg
    tags:
      chrony:
        chrony: imagehash
  discovery:
    kubernetesVersion: 1.20.5
    d8SpecificNodeCountByRole:
      worker: 3
      master: 3
`
	moduleValues = `
vpa:
  updateMode: "Auto"
  maxCPU: "50m"
  maxMemory: "100Mi"
ntpServers: ["pool.ntp.org", "ntp.ubuntu.com"]
`
)

var _ = Describe("Module :: chrony :: helm template ::", func() {
	f := SetupHelmConfig(``)

	Context("Render", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("global", globalValues)
			f.ValuesSetFromYaml("chrony", moduleValues)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			Expect(f.RenderError).ShouldNot(HaveOccurred())

			namespace := f.KubernetesGlobalResource("Namespace", "d8-chrony")
			registrySecret := f.KubernetesResource("Secret", "d8-chrony", "deckhouse-registry")

			chronyDaemonSetTest := f.KubernetesResource("DaemonSet", "d8-chrony", "chrony")
			chronyVPATest := f.KubernetesResource("VerticalPodAutoscaler", "d8-chrony", "chrony")
			chronyPDBTest := f.KubernetesResource("PodDisruptionBudget", "d8-chrony", "chrony")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(chronyDaemonSetTest.Exists()).To(BeTrue())
			Expect(chronyDaemonSetTest.Field("spec.template.spec.containers.0.env.0").String()).To(MatchJSON(`
  {
    "name": "NTP_SERVERS",
    "value": "pool.ntp.org ntp.ubuntu.com"
  }
`))

			Expect(chronyVPATest.Exists()).To(BeTrue())
			Expect(chronyVPATest.Field("spec.updatePolicy.updateMode").String()).To(Equal(`Auto`))
			Expect(chronyVPATest.Field("spec.resourcePolicy.containerPolicies.0.maxAllowed.cpu").String()).To(Equal(`50m`))
			Expect(chronyVPATest.Field("spec.resourcePolicy.containerPolicies.0.maxAllowed.memory").String()).To(Equal(`100Mi`))

			Expect(chronyPDBTest.Exists()).To(BeTrue())
		})
	})
})
