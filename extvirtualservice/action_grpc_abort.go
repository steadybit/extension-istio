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

type GrpcAbortAction struct {
}

func NewGrpcAbortAction() action_kit_sdk.Action[ActionState] {
	return GrpcAbortAction{}
}

var _ action_kit_sdk.Action[ActionState] = (*GrpcAbortAction)(nil)
var _ action_kit_sdk.ActionWithStop[ActionState] = (*GrpcAbortAction)(nil)

func (f GrpcAbortAction) NewEmptyState() ActionState {
	return ActionState{}
}

func (f GrpcAbortAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.grpc.abort", virtualServiceTargetID),
		Label:       "gRPC Abort",
		Description: "Injects a gRPC abort fault into all gRPC routes of the targeted virtual services. Abort requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
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
				Description:  extutil.Ptr("Duration defining for how long the gRPC abort should be injected."),
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
				Label:        "gRPC status code",
				Description:  extutil.Ptr("gRPC status code to use for aborted requests."),
				Type:         action_kit_api.String,
				DefaultValue: extutil.Ptr("UNAVAILABLE"),
				// See https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
				Options: extutil.Ptr([]action_kit_api.ParameterOption{
					action_kit_api.ExplicitParameterOption{
						Label: "Ok",
						Value: "OK",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Cancelled",
						Value: "CANCELLED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Unknown",
						Value: "UNKNOWN",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Invalid argument",
						Value: "INVALID_ARGUMENT",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Deadline exceeded",
						Value: "DEADLINE_EXCEEDED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Not found",
						Value: "NOT_FOUND",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Already exists",
						Value: "ALREADY_EXISTS",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Permission denied",
						Value: "PERMISSION_DENIED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Resource exhausted",
						Value: "RESOURCE_EXHAUSTED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Failed precondition",
						Value: "FAILED_PRECONDITION",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Aborted",
						Value: "ABORTED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Out of range",
						Value: "OUT_OF_RANGE",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Unimplemented",
						Value: "UNIMPLEMENTED",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Internal",
						Value: "INTERNAL",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Unavailable",
						Value: "UNAVAILABLE",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Data loss",
						Value: "DATA_LOSS",
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Unauthenticated",
						Value: "UNAUTHENTICATED",
					},
				}),
				Required: extutil.Ptr(true),
				Order:    extutil.Ptr(2),
			},
		}, getAdvancedTargetingParameters(3)...),
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f GrpcAbortAction) Prepare(_ context.Context, state *ActionState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	return nil, prepareVirtualServiceFault(state, request, toGrpcAbortFault)
}

func (f GrpcAbortAction) Start(ctx context.Context, state *ActionState) (*action_kit_api.StartResult, error) {
	return nil, startVirtualServiceFault(ctx, state)
}

func (f GrpcAbortAction) Stop(ctx context.Context, state *ActionState) (*action_kit_api.StopResult, error) {
	return nil, stopVirtualServiceFault(ctx, state)
}

func toGrpcAbortFault(request action_kit_api.PrepareActionRequestBody) *networkingv1beta1.HTTPFaultInjection {
	return &networkingv1beta1.HTTPFaultInjection{
		Abort: &networkingv1beta1.HTTPFaultInjection_Abort{
			ErrorType: &networkingv1beta1.HTTPFaultInjection_Abort_GrpcStatus{
				GrpcStatus: request.Config["statusCode"].(string),
			},
			Percentage: &networkingv1beta1.Percent{
				Value: request.Config["percentage"].(float64),
			},
		},
	}
}
