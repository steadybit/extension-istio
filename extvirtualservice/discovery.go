// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"context"
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-istio/extclient"
	"github.com/steadybit/extension-istio/extconfig"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"time"
)

const discoveryBasePath = basePath + "/discovery"

type serviceDiscovery struct {
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*serviceDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*serviceDiscovery)(nil)
)

func NewVirtualServiceDiscovery() discovery_kit_sdk.TargetDiscovery {
	discovery := &serviceDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 30*time.Second),
	)
}

func (d *serviceDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id: VirtualServiceTargetID,
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         discoveryBasePath + "/discovered-targets",
			CallInterval: extutil.Ptr("30s"),
		},
	}
}

func (d *serviceDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       VirtualServiceTargetID,
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

func (d *serviceDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
		{
			Attribute: "istio.virtual-service.name",
			Label: discovery_kit_api.PluralLabel{
				One:   "Virtual Service",
				Other: "Virtual Services",
			},
		},
	}
}

func (d *serviceDiscovery) DiscoverTargets(_ context.Context) ([]discovery_kit_api.Target, error) {
	return getVirtualServiceTargets(extclient.Istio), nil
}

func getVirtualServiceTargets(client *extclient.IstioClient) []discovery_kit_api.Target {
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
			TargetType: VirtualServiceTargetID,
			Attributes: attributes,
		}
	}

	return discovery_kit_commons.ApplyAttributeExcludes(result, extconfig.Config.DiscoveryAttributesExcludesVirtualSerice)
}
