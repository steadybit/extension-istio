// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extclient

import (
	"context"
	"errors"
	"flag"
	"github.com/rs/zerolog/log"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedClient "istio.io/client-go/pkg/clientset/versioned"
	informers "istio.io/client-go/pkg/informers/externalversions"
	"istio.io/client-go/pkg/listers/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

var Istio *IstioClient

type IstioClient struct {
	clientset               versionedClient.Interface
	virtualServicesLister   v1beta1.VirtualServiceLister
	virtualServicesInformer cache.SharedIndexInformer
}

func (c *IstioClient) GetVirtualServices() []*beta1.VirtualService {
	vs, err := c.virtualServicesLister.List(labels.Everything())

	if err != nil {
		log.Error().Err(err).Msgf("Failed fetching VirtualService resources")
		return []*beta1.VirtualService{}
	}

	return vs
}

func (c *IstioClient) AddHTTPFault(ctx context.Context, namespace string, name string, fault *networkingv1beta1.HTTPFaultInjection) error {
	return c.ModifyHTTPRoutes(ctx, namespace, name, func(http *networkingv1beta1.HTTPRoute) {
		http.Fault = fault.DeepCopy()
	})
}

func (c *IstioClient) RemoveAllFaults(ctx context.Context, namespace string, name string) error {
	return c.ModifyHTTPRoutes(ctx, namespace, name, func(http *networkingv1beta1.HTTPRoute) {
		http.Fault = nil
	})
}

func (c *IstioClient) ModifyHTTPRoutes(ctx context.Context, namespace string, name string, modifier func(route *networkingv1beta1.HTTPRoute)) error {
	vs, err := c.clientset.NetworkingV1beta1().VirtualServices(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return err
	}

	vs = vs.DeepCopy()
	for _, http := range vs.Spec.Http {
		modifier(http)
	}
	_, err = c.clientset.NetworkingV1beta1().VirtualServices(namespace).Update(ctx, vs, v1.UpdateOptions{})
	return err
}

func NewIstioClient(clientset versionedClient.Interface, stopCh <-chan struct{}) *IstioClient {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	virtualServices := factory.Networking().V1beta1().VirtualServices()
	virtualServicesInformer := virtualServices.Informer()

	go factory.Start(stopCh)

	log.Info().Msgf("Start Kubernetes cache sync.")
	if !cache.WaitForCacheSync(stopCh,
		virtualServicesInformer.HasSynced,
	) {
		log.Fatal().Msg("Timed out waiting for caches to sync")
	}
	log.Info().Msgf("Caches synced.")

	return &IstioClient{
		clientset:               clientset,
		virtualServicesLister:   virtualServices.Lister(),
		virtualServicesInformer: virtualServicesInformer,
	}
}

func PrepareClient(stopCh <-chan struct{}) {
	clientset := createIstioClientset()
	Istio = NewIstioClient(clientset, stopCh)
}

func createIstioClientset() versionedClient.Interface {
	config, err := rest.InClusterConfig()
	if err == nil {
		log.Info().Msgf("Extension is running inside a cluster, config found")
	} else if errors.Is(err, rest.ErrNotInCluster) {
		log.Info().Msgf("Extension is not running inside a cluster, try local .kube config")
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}

	if err != nil {
		log.Fatal().Err(err).Msgf("Could not find kubernetes config")
	}

	config.UserAgent = "steadybit-extension-kubernetes"
	config.Timeout = time.Second * 10

	istioClient, err := versionedClient.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create istio client")
	}

	info, err := istioClient.ServerVersion()
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not fetch server version.")
	}

	log.Info().Msgf("Cluster connected! Kubernetes Server Version %+v", info)

	return istioClient
}
