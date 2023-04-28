// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"reflect"
	"testing"
	"time"
)

func Test_toHTTPDelayFault(t *testing.T) {
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
				request: extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
					Config: map[string]interface{}{
						"delay":      5000.0,
						"percentage": 67.0,
					},
				}),
			},
			want: &networkingv1beta1.HTTPFaultInjection{
				Delay: &networkingv1beta1.HTTPFaultInjection_Delay{
					HttpDelayType: &networkingv1beta1.HTTPFaultInjection_Delay_FixedDelay{
						FixedDelay: durationpb.New(time.Second * 5),
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
			if got := toHTTPDelayFault(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toHTTPDelayFault() = %v, want %v", got, tt.want)
			}
		})
	}
}
