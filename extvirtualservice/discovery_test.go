// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"context"
	"github.com/steadybit/extension-istio/extclient"
	"github.com/steadybit/extension-istio/extconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedClient "istio.io/client-go/pkg/clientset/versioned"
	testclient "istio.io/client-go/pkg/clientset/versioned/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_getDiscoveredVirtualServices(t *testing.T) {
	// Given
	stopCh := make(chan struct{})
	defer close(stopCh)
	client, clientset := getTestClient(stopCh)
	extconfig.Config.ClusterName = "development"
	extconfig.Config.DiscoveryAttributesExcludesVirtualSerice = []string{"k8s.virtual-service.label.toIgnore"}

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
				Labels: map[string]string{
					"best-city": "Kevelaer",
					"toIgnore": "Bielefeld",
				},
			},
		}, v1.CreateOptions{})
	require.NoError(t, err)

	// When
	assert.Eventually(t, func() bool {
		return len(GetVirtualServiceTargets(client)) == 1
	}, time.Minute, 100*time.Millisecond)

	// Then
	targets := GetVirtualServiceTargets(client)
	require.Len(t, targets, 1)
	target := targets[0]
	require.Equal(t, "development/default/shop", target.Id)
	require.Equal(t, virtualServiceTargetID, target.TargetType)
	require.Equal(t, "shop", target.Label)
	require.Equal(t, map[string][]string{
		"istio.virtual-service.name":          {"shop"},
		"k8s.namespace":                       {"default"},
		"k8s.cluster-name":                    {"development"},
		"k8s.virtual-service.label.best-city": {"Kevelaer"},
	}, target.Attributes)
}

func getTestClient(stopCh <-chan struct{}) (*extclient.IstioClient, versionedClient.Interface) {
	clientset := testclient.NewSimpleClientset()
	client := extclient.NewIstioClient(clientset, stopCh)
	return client, clientset
}
