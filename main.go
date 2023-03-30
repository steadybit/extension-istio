// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-istio/extclient"
	"github.com/steadybit/extension-istio/extconfig"
	"github.com/steadybit/extension-istio/extvirtualservice"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
)

func main() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	extclient.PrepareClient(stopCh)

	extconfig.ParseConfiguration()
	extconfig.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))

	extvirtualservice.RegisterDiscoveryHandlers()
	extvirtualservice.RegisterHttpDelayActionHandlers()
	extvirtualservice.RegisterHttpAbortActionHandlers()
	extvirtualservice.RegisterGrpcAbortActionHandlers()

	exthttp.Listen(exthttp.ListenOpts{
		Port: 8080,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList       `json:",inline"`
	discovery_kit_api.DiscoveryList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList:    extvirtualservice.GetActionList(),
		DiscoveryList: extvirtualservice.GetDiscoveryList(),
	}
}
