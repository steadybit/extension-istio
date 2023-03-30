// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import "github.com/steadybit/action-kit/go/action_kit_api/v2"

func GetActionList() action_kit_api.ActionList {
	return action_kit_api.ActionList{
		Actions: []action_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   httpDelayActionBasePath,
			},
			{
				Method: "GET",
				Path:   httpAbortActionBasePath,
			},
			{
				Method: "GET",
				Path:   grpcAbortActionBasePath,
			},
		},
	}
}
