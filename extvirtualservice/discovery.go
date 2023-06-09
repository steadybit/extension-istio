// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-istio/extclient"
	"github.com/steadybit/extension-istio/extconfig"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
)

const discoveryBasePath = basePath + "/discovery"

func RegisterDiscoveryHandlers() {
	exthttp.RegisterHttpHandler(discoveryBasePath, exthttp.GetterAsHandler(getDiscoveryDescription))
	exthttp.RegisterHttpHandler(discoveryBasePath+"/target-description", exthttp.GetterAsHandler(getTargetDescription))
	exthttp.RegisterHttpHandler(discoveryBasePath+"/attribute-descriptions", exthttp.GetterAsHandler(getAttributeDescriptions))
	exthttp.RegisterHttpHandler(discoveryBasePath+"/discovered-targets", getDiscoveredTargets)
}

func GetDiscoveryList() discovery_kit_api.DiscoveryList {
	return discovery_kit_api.DiscoveryList{
		Discoveries: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   discoveryBasePath,
			},
		},
		TargetTypes: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   discoveryBasePath + "/target-description",
			},
		},
		TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   discoveryBasePath + "/attribute-descriptions",
			},
		},
	}
}

func getDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         virtualServiceTargetID,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         discoveryBasePath + "/discovered-targets",
			CallInterval: extutil.Ptr("30s"),
		},
	}
}

func getTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       virtualServiceTargetID,
		Icon:     extutil.Ptr(targetIcon),
		Label:    discovery_kit_api.PluralLabel{One: "Virtual Service", Other: "Virtual Services"},
		Category: extutil.Ptr("Kubernetes"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),

		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "istio.virtual-service.name"},
				{Attribute: "k8s.namespace"},
				{Attribute: "k8s.cluster-name"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "istio.virtual-service.name",
					Direction: "ASC",
				},
			},
		},
	}
}

func getAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{
		Attributes: []discovery_kit_api.AttributeDescription{
			{
				Attribute: "istio.virtual-service.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Virtual Service",
					Other: "Virtual Services",
				},
			},
		},
	}
}

func getDiscoveredTargets(w http.ResponseWriter, r *http.Request, _ []byte) {
	exthttp.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: GetVirtualServiceTargets(extclient.Istio)})
}

func GetVirtualServiceTargets(client *extclient.IstioClient) []discovery_kit_api.Target {
	virtualServices := client.GetVirtualServices()
	result := make([]discovery_kit_api.Target, len(virtualServices))

	for i, virtualService := range virtualServices {
		attributes := make(map[string][]string)
		attributes["istio.virtual-service.name"] = []string{virtualService.Name}
		attributes["k8s.namespace"] = []string{virtualService.Namespace}
		attributes["k8s.cluster-name"] = []string{extconfig.Config.ClusterName}

		for key, value := range virtualService.Labels {
			attributes["k8s.virtual-service.label."+key] = []string{value}
		}

		result[i] = discovery_kit_api.Target{
			Id:         fmt.Sprintf("%s/%s/%s", extconfig.Config.ClusterName, virtualService.Namespace, virtualService.Name),
			Label:      virtualService.Name,
			TargetType: virtualServiceTargetID,
			Attributes: attributes,
		}
	}

	return result
}
