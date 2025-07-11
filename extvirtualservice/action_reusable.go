// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-istio/extclient"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extutil"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
)

type ActionState struct {
	Namespace         string
	Name              string
	FaultyRoutePrefix string
	Fault             *networkingv1beta1.HTTPFaultInjection
	SourceLabels      map[string]string
	Headers           map[string]*networkingv1beta1.StringMatch
}

func getAdvancedTargetingParameters(startOrder int) []action_kit_api.ActionParameter {
	return []action_kit_api.ActionParameter{
		{
			Name:        "headers",
			Label:       "For requests with HTTP headers",
			Description: extutil.Ptr("Restrict the fault injection to those HTTP requests that carry all of these HTTP header key/value pairs."),
			Type:        action_kit_api.ActionParameterTypeKeyValue,
			Advanced:    extutil.Ptr(true),
			Required:    extutil.Ptr(false),
			Order:       extutil.Ptr(startOrder + 1),
		},
		{
			Name:        "headersMatchType",
			Label:       "HTTP header match type",
			Description: extutil.Ptr("How the header key/value pairs should be matched."),
			Type:        action_kit_api.ActionParameterTypeString,
			Options: extutil.Ptr([]action_kit_api.ParameterOption{
				action_kit_api.ExplicitParameterOption{
					Label: "Exact / equality",
					Value: "exact",
				},
				action_kit_api.ExplicitParameterOption{
					Label: "Prefix / starts with",
					Value: "prefix",
				},
				action_kit_api.ExplicitParameterOption{
					Label: "Regular expression (RE2 syntax)",
					Value: "regex",
				},
			}),
			DefaultValue: extutil.Ptr("exact"),
			Advanced:     extutil.Ptr(true),
			Required:     extutil.Ptr(true),
			Order:        extutil.Ptr(startOrder + 2),
		},
		{
			Name:        "sourceLabels",
			Label:       "For requests from sources labeled with",
			Description: extutil.Ptr("Restrict the fault injection to those HTTP requests coming from source (client) workloads with the given labels."),
			Type:        action_kit_api.ActionParameterTypeKeyValue,
			Advanced:    extutil.Ptr(true),
			Required:    extutil.Ptr(false),
			Order:       extutil.Ptr(startOrder + 3),
			Hint: extutil.Ptr(action_kit_api.ActionHint{
				Type:    action_kit_api.HintInfo,
				Content: "If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable.",
			}),
		},
	}
}

func prepareVirtualServiceFault(state *ActionState,
	request action_kit_api.PrepareActionRequestBody,
	toFault func(req action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection) error {

	headers, err := extutil.ToKeyValue(request.Config, "headers")
	if err != nil {
		return extension_kit.ToError("Failed prepare attack", err)
	}
	headersMatchType := request.Config["headersMatchType"].(string)

	headersWithMatchType := make(map[string]*networkingv1beta1.StringMatch, len(headers))
	for key, value := range headers {
		var match networkingv1beta1.StringMatch
		if headersMatchType == "prefix" {
			match = networkingv1beta1.StringMatch{
				MatchType: &networkingv1beta1.StringMatch_Prefix{
					Prefix: value,
				},
			}
		} else if headersMatchType == "regex" {
			match = networkingv1beta1.StringMatch{
				MatchType: &networkingv1beta1.StringMatch_Regex{
					Regex: value,
				},
			}
		} else {
			match = networkingv1beta1.StringMatch{
				MatchType: &networkingv1beta1.StringMatch_Exact{
					Exact: value,
				},
			}
		}

		headersWithMatchType[key] = &match
	}

	sourceLabels, err := extutil.ToKeyValue(request.Config, "sourceLabels")
	if err != nil {
		return extension_kit.ToError("Failed prepare attack", err)
	}

	state.Namespace = request.Target.Attributes["k8s.namespace"][0]
	state.Name = request.Target.Attributes["istio.virtual-service.name"][0]
	state.FaultyRoutePrefix = fmt.Sprintf("steadybit-injected-fault_%s", request.ExecutionId)
	state.Fault = toFault(request)
	state.Headers = headersWithMatchType
	state.SourceLabels = sourceLabels
	return nil
}

func startVirtualServiceFault(ctx context.Context, state *ActionState) error {
	err := extclient.Istio.AddHTTPFault(ctx, state.Namespace, state.Name, state.FaultyRoutePrefix, state.Fault, state.SourceLabels, state.Headers)
	if err != nil {
		return extension_kit.ToError(fmt.Sprintf("Failed to add HTTP fault to VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err)
	}
	return nil
}

func stopVirtualServiceFault(ctx context.Context, state *ActionState) error {
	err := extclient.Istio.RemoveAllFaults(ctx, state.Namespace, state.Name, state.FaultyRoutePrefix)
	if err != nil {
		return extension_kit.ToError(fmt.Sprintf("Failed to remove HTTP faults from VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err)
	}
	return nil
}
