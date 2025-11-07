// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	networkingv1 "istio.io/api/networking/v1"
	"reflect"
	"testing"
)

func Test_toGrpcAbortFault(t *testing.T) {
	type args struct {
		request action_kit_api.PrepareActionRequestBody
	}
	tests := []struct {
		name string
		args args
		want *networkingv1.HTTPFaultInjection
	}{
		{
			name: "generates a gRPC fault structure",
			args: args{
				request: extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
					Config: map[string]interface{}{
						"statusCode": "UNAVAILABLE",
						"percentage": 67.0,
					},
				}),
			},
			want: &networkingv1.HTTPFaultInjection{
				Abort: &networkingv1.HTTPFaultInjection_Abort{
					ErrorType: &networkingv1.HTTPFaultInjection_Abort_GrpcStatus{
						GrpcStatus: "UNAVAILABLE",
					},
					Percentage: &networkingv1.Percent{
						Value: 67.0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toGrpcAbortFault(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toGrpcAbortFault() = %v, want %v", got, tt.want)
			}
		})
	}
}
