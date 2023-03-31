// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"reflect"
	"testing"
)

func Test_toHTTPAbortFault(t *testing.T) {
	type args struct {
		request action_kit_api.PrepareActionRequestBody
	}
	tests := []struct {
		name string
		args args
		want *networkingv1beta1.HTTPFaultInjection
	}{
		{
			name: "generates a HTTP abort structure",
			args: args{
				request: action_kit_api.PrepareActionRequestBody{
					Config: map[string]interface{}{
						"statusCode": 404.0,
						"percentage": 67.0,
					},
				},
			},
			want: &networkingv1beta1.HTTPFaultInjection{
				Abort: &networkingv1beta1.HTTPFaultInjection_Abort{
					ErrorType: &networkingv1beta1.HTTPFaultInjection_Abort_HttpStatus{
						HttpStatus: 404,
					},
					Percentage: &networkingv1beta1.Percent{
						Value: 67.0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toHTTPAbortFault(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toHTTPAbortFault() = %v, want %v", got, tt.want)
			}
		})
	}
}
