// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"net/http"
)

const httpAbortActionBasePath = basePath + "/actions/http-abort"

func RegisterHttpAbortActionHandlers() {
	exthttp.RegisterHttpHandler(httpAbortActionBasePath, exthttp.GetterAsHandler(getHttpAbortActionDescription))
	exthttp.RegisterHttpHandler(httpAbortActionBasePath+"/prepare", prepareHttpAbort)
	exthttp.RegisterHttpHandler(httpAbortActionBasePath+"/start", startVirtualServiceFault)
	exthttp.RegisterHttpHandler(httpAbortActionBasePath+"/stop", stopVirtualServiceFault)
}

func getHttpAbortActionDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.http.abort", virtualServiceTargetId),
		Label:       "HTTP Abort",
		Description: "Injects a HTTP abort fault into all HTTP routes of the targeted virtual services. Abort requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
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
				Description:  extutil.Ptr("Duration defining for how long the HTTP abort should be injected."),
				Type:         action_kit_api.Duration,
				DefaultValue: extutil.Ptr("30s"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(0),
			},
			{
				Name:         "percentage",
				Label:        "Percentage",
				Description:  extutil.Ptr("Percentage of requests on which the abort will be injected."),
				Type:         action_kit_api.Duration,
				DefaultValue: extutil.Ptr("50"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(1),
			},
			{
				Name:         "statusCode",
				Label:        "HTTP Status Code",
				Description:  extutil.Ptr("HTTP status code to use for aborted requests."),
				Type:         action_kit_api.Integer,
				DefaultValue: extutil.Ptr("500"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(2),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpAbortActionBasePath + "/prepare",
		},
		Start: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpAbortActionBasePath + "/start",
		},
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   httpAbortActionBasePath + "/stop",
		}),
	}
}

func prepareHttpAbort(w http.ResponseWriter, r *http.Request, body []byte) {
	prepareVirtualServiceFault(w, r, body, toHttpAbortFault)
}

func toHttpAbortFault(request action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection {
	return &networkingv1beta1.HTTPFaultInjection{
		Abort: &networkingv1beta1.HTTPFaultInjection_Abort{
			ErrorType: &networkingv1beta1.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: int32(request.Config["statusCode"].(float64)),
			},
			Percentage: &networkingv1beta1.Percent{
				Value: request.Config["percentage"].(float64),
			},
		},
	}
}
