// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"context"
	"github.com/google/uuid"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-istio/extclient"
	"github.com/steadybit/extension-istio/extconfig"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/require"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_attackLifecycle(t *testing.T) {
	// General preparation
	stopCh := make(chan struct{})
	defer close(stopCh)
	client, clientset := getTestClient(stopCh)
	extclient.Istio = client
	extconfig.Config.ClusterName = "development"

	_, err := clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Create(context.Background(), &v1beta1.VirtualService{
			TypeMeta: v1.TypeMeta{
				Kind:       "VirtualService",
				APIVersion: "v1beta1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "shop",
				Namespace: "default",
			},
			Spec: networkingv1beta1.VirtualService{
				Http: []*networkingv1beta1.HTTPRoute{
					{
						Name: "test-route-1",
						Match: []*networkingv1beta1.HTTPMatchRequest{
							{
								Headers: map[string]*networkingv1beta1.StringMatch{
									"Content-Type": {
										MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
									},
								},
								SourceLabels: map[string]string{
									"env": "prod",
								},
							},
						},
					},
					{
						Name: "test-route-2",
					},
				},
			},
		}, v1.CreateOptions{})
	require.NoError(t, err)

	// Prepare call
	prepareRequest := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		ExecutionId: uuid.MustParse("22955847-b455-461d-8f9b-61ef1ef05060"),
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"k8s.namespace":              {"default"},
				"istio.virtual-service.name": {"shop"},
			},
		},
		Config: map[string]interface{}{
			"delay":            5000.0,
			"percentage":       69.0,
			"sourceLabels":     []any{},
			"headers":          []any{},
			"headersMatchType": "exact",
		},
	})
	state := ActionState{}
	err = prepareVirtualServiceFault(&state, prepareRequest, toHTTPDelayFault)
	require.NoError(t, err)

	// Start call
	err = startVirtualServiceFault(context.TODO(), &state)
	require.NoError(t, err)

	// Check that the VirtualService has a configured fault
	vs, err := clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Get(context.Background(), "shop", v1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, vs.Spec.Http, 4)
	require.Equal(t, "steadybit-injected-fault_22955847-b455-461d-8f9b-61ef1ef05060_0", vs.Spec.Http[0].Name)
	require.NotNil(t, vs.Spec.Http[0].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
			},
		},
	}, vs.Spec.Http[0].Match)
	require.Equal(t, "test-route-1", vs.Spec.Http[1].Name)
	require.Nil(t, vs.Spec.Http[1].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
			},
		},
	}, vs.Spec.Http[1].Match)
	require.Equal(t, "steadybit-injected-fault_22955847-b455-461d-8f9b-61ef1ef05060_1", vs.Spec.Http[2].Name)
	require.NotNil(t, vs.Spec.Http[2].Fault)
	require.Len(t, vs.Spec.Http[2].Match, 0)
	require.Equal(t, "test-route-2", vs.Spec.Http[3].Name)
	require.Nil(t, vs.Spec.Http[3].Fault)
	require.Len(t, vs.Spec.Http[3].Match, 0)

	// Stop call
	err = stopVirtualServiceFault(context.TODO(), &state)
	require.NoError(t, err)

	// Check that the faults were removed from the VirtualService resource
	vs, err = clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Get(context.Background(), "shop", v1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, vs.Spec.Http, 2)
	require.Equal(t, "test-route-1", vs.Spec.Http[0].Name)
	require.Nil(t, vs.Spec.Http[0].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
			},
		},
	}, vs.Spec.Http[0].Match)
	require.Equal(t, "test-route-2", vs.Spec.Http[1].Name)
	require.Nil(t, vs.Spec.Http[1].Fault)
	require.Len(t, vs.Spec.Http[1].Match, 0)
}

func Test_attackLifecycle_with_client_restriction(t *testing.T) {
	// General preparation
	stopCh := make(chan struct{})
	defer close(stopCh)
	client, clientset := getTestClient(stopCh)
	extclient.Istio = client
	extconfig.Config.ClusterName = "development"

	_, err := clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Create(context.Background(), &v1beta1.VirtualService{
			TypeMeta: v1.TypeMeta{
				Kind:       "VirtualService",
				APIVersion: "v1beta1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "shop",
				Namespace: "default",
			},
			Spec: networkingv1beta1.VirtualService{
				Http: []*networkingv1beta1.HTTPRoute{
					{
						Name: "test-route-1",
						Match: []*networkingv1beta1.HTTPMatchRequest{
							{
								Headers: map[string]*networkingv1beta1.StringMatch{
									"Content-Type": {
										MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
									},
								},
								SourceLabels: map[string]string{
									"env": "prod",
								},
							},
						},
					},
					{
						Name: "test-route-2",
					},
				},
			},
		}, v1.CreateOptions{})
	require.NoError(t, err)

	// Prepare call
	prepareRequest := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		ExecutionId: uuid.MustParse("22955847-b455-461d-8f9b-61ef1ef05060"),
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"k8s.namespace":              {"default"},
				"istio.virtual-service.name": {"shop"},
			},
		},
		Config: map[string]interface{}{
			"delay":      5000.0,
			"percentage": 69.0,
			"sourceLabels": []interface{}{
				map[string]interface{}{"key": "app", "value": "shop"},
			},
			"headers": []interface{}{
				map[string]interface{}{"key": "Accept", "value": "application/json"},
			},
			"headersMatchType": "exact",
		},
	})
	state := ActionState{}
	err = prepareVirtualServiceFault(&state, prepareRequest, toHTTPDelayFault)
	require.NoError(t, err)

	// Start call
	err = startVirtualServiceFault(context.TODO(), &state)
	require.NoError(t, err)

	// Check that the VirtualService has a configured fault
	vs, err := clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Get(context.Background(), "shop", v1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, vs.Spec.Http, 4)
	require.Equal(t, "steadybit-injected-fault_22955847-b455-461d-8f9b-61ef1ef05060_0", vs.Spec.Http[0].Name)
	require.NotNil(t, vs.Spec.Http[0].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
				"Accept": {
					MatchType: &networkingv1beta1.StringMatch_Exact{Exact: "application/json"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
				"app": "shop",
			},
		},
	}, vs.Spec.Http[0].Match)
	require.Equal(t, "test-route-1", vs.Spec.Http[1].Name)
	require.Nil(t, vs.Spec.Http[1].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
			},
		},
	}, vs.Spec.Http[1].Match)
	require.Equal(t, "steadybit-injected-fault_22955847-b455-461d-8f9b-61ef1ef05060_1", vs.Spec.Http[2].Name)
	require.NotNil(t, vs.Spec.Http[2].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Accept": {
					MatchType: &networkingv1beta1.StringMatch_Exact{Exact: "application/json"},
				},
			},
			SourceLabels: map[string]string{
				"app": "shop",
			},
		},
	}, vs.Spec.Http[2].Match)
	require.Equal(t, "test-route-2", vs.Spec.Http[3].Name)
	require.Nil(t, vs.Spec.Http[3].Fault)
	require.Len(t, vs.Spec.Http[3].Match, 0)

	// Stop call
	err = stopVirtualServiceFault(context.TODO(), &state)
	require.NoError(t, err)

	// Check that the faults were removed from the VirtualService resource
	vs, err = clientset.
		NetworkingV1beta1().
		VirtualServices("default").
		Get(context.Background(), "shop", v1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, vs.Spec.Http, 2)
	require.Equal(t, "test-route-1", vs.Spec.Http[0].Name)
	require.Nil(t, vs.Spec.Http[0].Fault)
	require.Equal(t, []*networkingv1beta1.HTTPMatchRequest{
		{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"Content-Type": {
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: "text/.*"},
				},
			},
			SourceLabels: map[string]string{
				"env": "prod",
			},
		},
	}, vs.Spec.Http[0].Match)
	require.Equal(t, "test-route-2", vs.Spec.Http[1].Name)
	require.Nil(t, vs.Spec.Http[1].Fault)
	require.Len(t, vs.Spec.Http[1].Match, 0)
}
