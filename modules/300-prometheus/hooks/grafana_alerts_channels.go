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
	"fmt"
	"time"

	"github.com/deckhouse/deckhouse/go_lib/module"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const madisonAlertChannelName = "madison"

const alertManagerGrafanaAlertChannelType = "prometheus-alertmanager"

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/prometheus/grafana_alerts_channels",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "grafana_alerts_channels",
			ApiVersion: "deckhouse.io/v1alpha1",
			Kind:       "GrafanaAlertsChannel",
			FilterFunc: filterGrafanaAlertsChannelCRD,
		},
	},
}, grafanaAlertsChannelsHandler)

type GrafanaAlertsChannel struct {
	OrgID                 int                    `json:"org_id"`
	Type                  string                 `json:"type"`
	Name                  string                 `json:"name"`
	UID                   string                 `json:"uid"`
	IsDefault             bool                   `json:"is_default"`
	DisableResolveMessage bool                   `json:"disable_resolve_message"`
	SendReminder          bool                   `json:"send_reminder"`
	Frequency             time.Duration          `json:"frequency,omitempty"`
	Settings              map[string]interface{} `json:"settings"`
	SecureSettings        map[string]interface{} `json:"secure_settings"`
}

func getStringFromUnstructured(obj *unstructured.Unstructured, path string) (string, error) {
	val, ok, err := unstructured.NestedString(obj.Object, path)
	if err != nil {
		return "", fmt.Errorf("cannot get '%s' from GrafanaNotificationsChannel: %v", path, err)
	}
	if !ok {
		return "", fmt.Errorf("has no '%s' field in GrafanaNotificationsChannel", path)
	}

	return val, nil
}

func getChannelSettings(notifierType string, obj *unstructured.Unstructured) (s map[string]interface{}, sec map[string]interface{}, err error) {
	if notifierType != alertManagerGrafanaAlertChannelType {
		return nil, nil, fmt.Errorf("unsupported GrafanaNotificationsChannel type %s", notifierType)
	}

	address, err := getStringFromUnstructured(obj, "spec.alertManager.address")
	if err != nil {
		return nil, nil, err
	}

	settings := map[string]interface{}{
		"url": address,
	}

	secureSettings := make(map[string]interface{})

	const authPath = "spec.alertManager.auth.basic"

	auth, ok, err := unstructured.NestedMap(obj.Object, authPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get '%s' from GrafanaNotificationsChannel: %v", authPath, err)
	}

	if ok {
		settings["basicAuthUser"] = auth["username"].(string)
		secureSettings["basicAuthPassword"] = auth["password"].(string)
	}

	// url can be without auth

	return settings, secureSettings, nil
}

func filterGrafanaAlertsChannelCRD(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	chType, err := getStringFromUnstructured(obj, "spec.type")
	if err != nil {
		return nil, fmt.Errorf("cannot get spec.type from GrafanaNotificationsChannel: %v", err)
	}

	disableResolveMsg, ok, err := unstructured.NestedBool(obj.Object, "spec.disableResolveMessage")
	if err != nil {
		return nil, fmt.Errorf("cannot get spec.disableResolveMessage from GrafanaNotificationsChannel: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("has no spec.disableResolveMessage field in GrafanaNotificationsChannel")
	}

	settings, securitySettings, err := getChannelSettings(chType, obj)
	if err != nil {
		return nil, err
	}

	return &GrafanaAlertsChannel{
		OrgID:                 1,
		Name:                  obj.GetName(),
		UID:                   obj.GetName(),
		Type:                  chType,
		DisableResolveMessage: disableResolveMsg,
		Settings:              settings,
		SecureSettings:        securitySettings,
	}, nil

}

func grafanaAlertsChannelsHandler(input *go_hook.HookInput) error {
	alertsChannelsRaw := input.Snapshots["grafana_alerts_channels"]

	alertsChannels := make([]*GrafanaAlertsChannel, 0)

	if module.IsEnabled("flant-integration", input) {
		clusterDomain := input.Values.Get("global.discovery.clusterDomain").String()

		// todo maybe get port from service? But we get unnecessary subscription to api server
		// and port will not change in future
		madisonProxyUrl := fmt.Sprintf("http://madison-proxy.d8-monitoring.svc.%s:8080", clusterDomain)

		alertsChannels = append(alertsChannels, &GrafanaAlertsChannel{
			OrgID:                 1,
			Name:                  madisonAlertChannelName,
			UID:                   madisonAlertChannelName,
			Type:                  alertManagerGrafanaAlertChannelType,
			DisableResolveMessage: false,
			Settings: map[string]interface{}{
				"url": madisonProxyUrl,
			},
			SecureSettings: make(map[string]interface{}),
		})
	}

	for _, nchRaw := range alertsChannelsRaw {
		nch := nchRaw.(*GrafanaAlertsChannel)
		alertsChannels = append(alertsChannels, nch)
	}

	input.Values.Set("prometheus.internal.grafana.alertsChannels", alertsChannels)

	return nil
}
