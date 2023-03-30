// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"net/http"
	"time"
)

const httpDelayActionBasePath = basePath + "/actions/http-delay"

func RegisterHTTPDelayActionHandlers() {
	exthttp.RegisterHttpHandler(httpDelayActionBasePath, exthttp.GetterAsHandler(getHTTPDelayActionDescription))
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/prepare", prepareHTTPDelay)
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/start", startVirtualServiceFault)
	exthttp.RegisterHttpHandler(httpDelayActionBasePath+"/stop", stopVirtualServiceFault)
}

func getHTTPDelayActionDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.http.delay", virtualServiceTargetID),
		Label:       "HTTP Delay",
		Description: "Injects a HTTP delay fault into all HTTP routes of the targeted virtual services. Delay requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetType:  extutil.Ptr(virtualServiceTargetID),
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
				Type:         action_kit_api.Percentage,
				DefaultValue: extutil.Ptr("50"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(1),
			},
			{
				Name:         "delay",
				Label:        "Delay",
				Description:  extutil.Ptr("Fixed delay before forwarding the request."),
				Type:         action_kit_api.Duration,
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

func prepareHTTPDelay(w http.ResponseWriter, r *http.Request, body []byte) {
	prepareVirtualServiceFault(w, r, body, toHTTPDelayFault)
}

func toHTTPDelayFault(request action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection {
	return &networkingv1beta1.HTTPFaultInjection{
		Delay: &networkingv1beta1.HTTPFaultInjection_Delay{
			HttpDelayType: &networkingv1beta1.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: durationpb.New(time.Millisecond * time.Duration(request.Config["delay"].(float64))),
			},
			Percentage: &networkingv1beta1.Percent{
				Value: request.Config["percentage"].(float64),
			},
		},
	}
}
