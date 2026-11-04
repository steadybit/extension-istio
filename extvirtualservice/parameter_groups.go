// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extvirtualservice

import "github.com/steadybit/action-kit/go/action_kit_api/v2"

const groupTrafficFilter = "Traffic Filter"

var trafficFilterGroup = action_kit_api.ParameterGroup{
	Name:        groupTrafficFilter,
	Description: "Leave unchanged to affect all requests to the selected virtual services.",
}
