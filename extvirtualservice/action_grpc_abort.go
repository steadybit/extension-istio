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

const grpcAbortActionBasePath = basePath + "/actions/grpc-abort"

func RegisterGrpcAbortActionHandlers() {
	exthttp.RegisterHttpHandler(grpcAbortActionBasePath, exthttp.GetterAsHandler(getGrpcAbortActionDescription))
	exthttp.RegisterHttpHandler(grpcAbortActionBasePath+"/prepare", prepareGrpcAbort)
	exthttp.RegisterHttpHandler(grpcAbortActionBasePath+"/start", startVirtualServiceFault)
	exthttp.RegisterHttpHandler(grpcAbortActionBasePath+"/stop", stopVirtualServiceFault)
}

func getGrpcAbortActionDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.grpc.abort", virtualServiceTargetID),
		Label:       "gRPC Abort",
		Description: "Injects a gRPC abort fault into all gRPC routes of the targeted virtual services. Abort requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc.",
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
		Prepare: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   grpcAbortActionBasePath + "/prepare",
		},
		Start: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   grpcAbortActionBasePath + "/start",
		},
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   grpcAbortActionBasePath + "/stop",
		}),
	}
}

func prepareGrpcAbort(w http.ResponseWriter, r *http.Request, body []byte) {
	prepareVirtualServiceFault(w, r, body, toGrpcAbortFault)
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
