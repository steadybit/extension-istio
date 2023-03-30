// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"encoding/json"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-istio/extclient"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"net/http"
	"time"
)

const httpDelayActionBasePath = basePath + "/actions/http-delay"

func RegisterActionHandlers() {
	exthttp.RegisterHttpHandler(httpDelayActionBasePath, exthttp.GetterAsHandler(getHttpDelayActionDescription))
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/prepare", prepareHttpDelay)
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/start", startHttpDelay)
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/stop", stopHttpDelay)
}

func getHttpDelayActionDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.http.delay", virtualServiceTargetId),
		Label:       "HTTP Delay",
		Description: "Injects a HTTP delay fault into all HTTP routes of the targeted virtual services. Delay requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetType:  extutil.Ptr(virtualServiceTargetId),
		TargetSelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
			{
				Label: "by name",
				Query: "istio.virtual-service.name=\"\"",
			},
		}),
		Category:    extutil.Ptr("Istio"),
		Kind:        action_kit_api.Attack,
		TimeControl: action_kit_api.External,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "duration",
				Label:        "Duration",
				Description:  extutil.Ptr("Duration defining for how long the HTTP delay should be injected."),
				Type:         action_kit_api.Duration,
				DefaultValue: extutil.Ptr("30s"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(0),
			},
			{
				Name:         "percentage",
				Label:        "Percentage",
				Description:  extutil.Ptr("Percentage of requests on which the delay will be injected."),
				Type:         action_kit_api.Duration,
				DefaultValue: extutil.Ptr("50"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(1),
			},
			{
				Name:         "delay",
				Label:        "Delay",
				Description:  extutil.Ptr("Fixed delay before forwarding the request."),
				Type:         action_kit_api.Percentage,
				DefaultValue: extutil.Ptr("5s"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(2),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpDelayActionBasePath + "/prepare",
		},
		Start: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpDelayActionBasePath + "/start",
		},
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpDelayActionBasePath + "/stop",
		}),
	}
}

type HttpDelayActionState struct {
	Namespace  string
	Name       string
	Percentage float64
	Delay      time.Duration
}

func prepareHttpDelay(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request action_kit_api.PrepareActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	state := HttpDelayActionState{
		Namespace:  request.Target.Attributes["k8s.namespace"][0],
		Name:       request.Target.Attributes["istio.virtual-service.name"][0],
		Percentage: request.Config["percentage"].(float64),
		Delay:      time.Millisecond * time.Duration(request.Config["delay"].(float64)),
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

func startHttpDelay(w http.ResponseWriter, r *http.Request, body []byte) {
	var request action_kit_api.StartActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state HttpDelayActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert action state", err))
		return
	}

	fault := networkingv1beta1.HTTPFaultInjection{
		Delay: &networkingv1beta1.HTTPFaultInjection_Delay{
			HttpDelayType: &networkingv1beta1.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: durationpb.New(state.Delay),
			},
			Percentage: &networkingv1beta1.Percent{
				Value: state.Percentage,
			},
		},
	}

	err = extclient.Istio.AddHttpFault(r.Context(), state.Namespace, state.Name, &fault)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError(fmt.Sprintf("Failed to add HTTP faults to VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.StartResult{})
}

func stopHttpDelay(w http.ResponseWriter, r *http.Request, body []byte) {
	var request action_kit_api.StopActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state HttpDelayActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert action state", err))
		return
	}

	err = extclient.Istio.RemoveAllFaults(r.Context(), state.Namespace, state.Name)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError(fmt.Sprintf("Failed to remove HTTP faults from VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.StopResult{})
}
