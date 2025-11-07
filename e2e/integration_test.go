// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	actValidate "github.com/steadybit/action-kit/go/action_kit_test/validate"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	disValidate "github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-istio/extvirtualservice"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networking "istio.io/api/networking/v1"
	apiv1 "istio.io/client-go/pkg/apis/networking/v1"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os/exec"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()

	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-istio",
		Port: 8080,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "logging.level=debug",
				"--set", "kubernetes.clusterName=minikube",
			}
		},
	}

	mOpts := e2e.DefaultMinikubeOpts().
		WithRuntimes(e2e.RuntimeDocker).
		AfterStart(func(m *e2e.Minikube) error {
			clientset, err := versioned.NewForConfig(m.ClientConfig)
			assert.NoError(t, err)
			err = exec.CommandContext(context.Background(), "minikube", []string{"-p", m.Profile, "addons", "enable", "istio-provisioner"}...).Run()
			assert.NoError(t, err)
			err = exec.CommandContext(context.Background(), "minikube", []string{"-p", m.Profile, "addons", "enable", "istio"}...).Run()
			assert.NoError(t, err)
			time.Sleep(30 * time.Second) //ToDo replace and wait for istio to be ready (namespace istio-system)

			httpRouteDest := &networking.HTTPRouteDestination{
				Destination: &networking.Destination{
					Host: "host.minikube.internal",
				},
				Headers: &networking.Headers{
					Request: &networking.Headers_HeaderOperations{
						Set: map[string]string{"Host": "host.minikube.internal"},
					},
				},
				Weight: 100,
			}
			_, err = clientset.
				NetworkingV1().
				VirtualServices("default").
				Create(context.Background(), &apiv1.VirtualService{
					Spec: networking.VirtualService{
						Hosts:    []string{"host.minikube.internal"},
						Gateways: nil,
						Http: []*networking.HTTPRoute{
							{
								Match: nil,
								Route: []*networking.HTTPRouteDestination{
									httpRouteDest,
								},
							},
						},
						Tls:      nil,
						Tcp:      nil,
						ExportTo: nil,
					},
					TypeMeta: v1.TypeMeta{
						Kind:       "VirtualService",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "shop",
						Namespace: "default",
						Labels: map[string]string{
							"best-city": "Kevelaer",
						},
					},
				}, v1.CreateOptions{})
			if err != nil {
				log.Err(err).Msg("Failed to create VirtualService")
			}
			require.NoError(t, err)

			return nil
		})

	e2e.WithMinikube(t, mOpts, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "validate discovery",
			Test: validateDiscovery,
		},
		{
			Name: "target discovery",
			Test: testDiscovery,
		},
		{
			Name: "validate Actions",
			Test: validateActions,
		},
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, disValidate.ValidateEndpointReferences("/", e.Client))
}

func testDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testDiscovery")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, extvirtualservice.VirtualServiceTargetID, func(target discovery_kit_api.Target) bool {
		log.Info().Msgf("Checking target: %v", target)
		return e2e.HasAttribute(target, "istio.virtual-service.name", "shop")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extvirtualservice.VirtualServiceTargetID)
	assert.Contains(t, target.Attributes, "istio.virtual-service.name")
	assert.Contains(t, target.Attributes, "k8s.namespace")
	assert.Contains(t, target.Attributes, "k8s.cluster-name")
	assert.Contains(t, target.Attributes, "k8s.virtual-service.label.best-city")
	assert.True(t, e2e.HasAttribute(target, "istio.virtual-service.name", "shop"))
	assert.True(t, e2e.HasAttribute(target, "k8s.namespace", "default"))
	assert.True(t, e2e.HasAttribute(target, "k8s.cluster-name", "minikube"))
	assert.True(t, e2e.HasAttribute(target, "k8s.virtual-service.label.best-city", "Kevelaer"))
}

func validateActions(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, actValidate.ValidateEndpointReferences("/", e.Client))
}
