// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
)

type HttpAbortAction struct {
}

func NewHttpAbortAction() action_kit_sdk.Action[ActionState] {
	return HttpAbortAction{}
}

var _ action_kit_sdk.Action[ActionState] = (*HttpAbortAction)(nil)
var _ action_kit_sdk.ActionWithStop[ActionState] = (*HttpAbortAction)(nil)

func (f HttpAbortAction) NewEmptyState() ActionState {
	return ActionState{}
}

func (f HttpAbortAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.http.abort", virtualServiceTargetID),
		Label:       "HTTP Abort",
		Description: "Injects a HTTP abort fault into all HTTP routes of the targeted virtual services. Abort requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: virtualServiceTargetID,
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label: "by name",
					Query: "istio.virtual-service.name=\"\"",
				},
			}),
		}),
		Category:    extutil.Ptr("Istio"),
		Kind:        action_kit_api.Attack,
		TimeControl: action_kit_api.External,
		Parameters: append([]action_kit_api.ActionParameter{
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
				Type:         action_kit_api.Percentage,
				DefaultValue: extutil.Ptr("50"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(1),
			},
			{
				Name:         "statusCode",
				Label:        "HTTP status code",
				Description:  extutil.Ptr("HTTP status code to use for aborted requests."),
				Type:         action_kit_api.Integer,
				DefaultValue: extutil.Ptr("500"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(2),
			},
		}, getAdvancedTargetingParameters(3)...),
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f HttpAbortAction) Prepare(_ context.Context, state *ActionState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	return nil, prepareVirtualServiceFault(state, request, toHTTPAbortFault)
}

func (f HttpAbortAction) Start(ctx context.Context, state *ActionState) (*action_kit_api.StartResult, error) {
	return nil, startVirtualServiceFault(ctx, state)
}

func (f HttpAbortAction) Stop(ctx context.Context, state *ActionState) (*action_kit_api.StopResult, error) {
	return nil, stopVirtualServiceFault(ctx, state)
}

func toHTTPAbortFault(request action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection {
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
