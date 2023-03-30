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
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"net/http"
)

type ActionState struct {
	Namespace string
	Name      string
	Fault     *networkingv1beta1.HTTPFaultInjection
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

	state := ActionState{
		Namespace: request.Target.Attributes["k8s.namespace"][0],
		Name:      request.Target.Attributes["istio.virtual-service.name"][0],
		Fault:     toFault(request),
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

	err = extclient.Istio.AddHTTPFault(r.Context(), state.Namespace, state.Name, state.Fault)
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

	err = extclient.Istio.RemoveAllFaults(r.Context(), state.Namespace, state.Name)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError(fmt.Sprintf("Failed to remove HTTP faults from VirtualService %s in namespace %s through Kubernetes API.", state.Name, state.Namespace), err))
		return
	}

	exthttp.WriteBody(w, action_kit_api.StopResult{})
}
