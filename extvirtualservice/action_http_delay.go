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
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"time"
)

type HttpDelayAction struct {
}

func NewHttpDelayAction() action_kit_sdk.Action[ActionState] {
	return HttpDelayAction{}
}

var _ action_kit_sdk.Action[ActionState] = (*HttpDelayAction)(nil)
var _ action_kit_sdk.ActionWithStop[ActionState] = (*HttpDelayAction)(nil)

func (f HttpDelayAction) NewEmptyState() ActionState {
	return ActionState{}
}

func (f HttpDelayAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.http.delay", VirtualServiceTargetID),
		Label:       "HTTP Delay",
		Description: "Injects a HTTP delay fault into all HTTP routes of the targeted virtual services. Delay requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: VirtualServiceTargetID,
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label: "by name",
					Query: "istio.virtual-service.name=\"\"",
				},
			}),
		}),
		Technology:  extutil.Ptr("Istio"),
		Kind:        action_kit_api.Attack,
		TimeControl: action_kit_api.TimeControlExternal,
		Parameters: append([]action_kit_api.ActionParameter{
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
		}, getAdvancedTargetingParameters(3)...),
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f HttpDelayAction) Prepare(_ context.Context, state *ActionState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	return nil, prepareVirtualServiceFault(state, request, toHTTPDelayFault)
}

func (f HttpDelayAction) Start(ctx context.Context, state *ActionState) (*action_kit_api.StartResult, error) {
	return nil, startVirtualServiceFault(ctx, state)
}

func (f HttpDelayAction) Stop(ctx context.Context, state *ActionState) (*action_kit_api.StopResult, error) {
	return nil, stopVirtualServiceFault(ctx, state)
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
