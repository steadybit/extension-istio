// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"encoding/json"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-istio/extclient"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"net/http"
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
			Type:        action_kit_api.KeyValue,
			Advanced:    extutil.Ptr(true),
			Required:    extutil.Ptr(false),
			Order:       extutil.Ptr(startOrder + 1),
		},
		{
			Name:        "headersMatchType",
			Label:       "HTTP header match type",
			Description: extutil.Ptr("How the header key/value pairs should be matched."),
			Type:        action_kit_api.String,
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
			Type:        action_kit_api.KeyValue,
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

func prepareVirtualServiceFault(w http.ResponseWriter,
	_ *http.Request,
	body []byte,
	toFault func(req action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection) {
	var request action_kit_api.PrepareActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	headers, err := toKeyValue(request, "headers")
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed prepare attack", err))
		return
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

	sourceLabels, err := toKeyValue(request, "sourceLabels")
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed prepare attack", err))
		return
	}

	state := ActionState{
		Namespace:         request.Target.Attributes["k8s.namespace"][0],
		Name:              request.Target.Attributes["istio.virtual-service.name"][0],
		FaultyRoutePrefix: fmt.Sprintf("steadybit-injected-fault_%s", request.ExecutionId),
		Fault:             toFault(request),
		Headers:           headersWithMatchType,
		SourceLabels:      sourceLabels,
	}

	var convertedState action_kit_api.ActionState
	err = extconversion.Convert(state, &convertedState)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode action state", err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.PrepareResult{
		State: convertedState,
	})
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func toKeyValue(request action_kit_api.PrepareActionRequestBody, configName string) (map[string]string, error) {
	kv, ok := request.Config[configName].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to interpret config value for %s as a key/value array", configName)
	}

	result := make(map[string]string, len(kv))
	for _, rawEntry := range kv {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("failed to interpret config value for %s as a key/value array", configName)
		}
		result[entry["key"].(string)] = entry["value"].(string)
	}

	return result, nil
}

func startVirtualServiceFault(w http.ResponseWriter, r *http.Request, body []byte) {
	var request action_kit_api.StartActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state ActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert action state", err))
		return
	}

	err = extclient.Istio.AddHTTPFault(r.Context(), state.Namespace, state.Name, state.FaultyRoutePrefix, state.Fault, state.SourceLabels, state.Headers)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError(fmt.Sprintf("Failed to add HTTP fault to VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.StartResult{})
}

func stopVirtualServiceFault(w http.ResponseWriter, r *http.Request, body []byte) {
	var request action_kit_api.StopActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state ActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert action state", err))
		return
	}

	err = extclient.Istio.RemoveAllFaults(r.Context(), state.Namespace, state.Name, state.FaultyRoutePrefix)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError(fmt.Sprintf("Failed to remove HTTP faults from VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.StopResult{})
}
