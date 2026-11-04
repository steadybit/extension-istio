// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extvirtualservice

import (
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
)

func TestParameterGroups(t *testing.T) {
	for name, desc := range map[string]action_kit_api.ActionDescription{
		"httpAbort": NewHttpAbortAction().Describe(),
		"httpDelay": NewHttpDelayAction().Describe(),
		"grpcAbort": NewGrpcAbortAction().Describe(),
	} {
		t.Run(name, func(t *testing.T) {
			var grouped []string
			lastOrderOutside := -1
			firstOrderInside := 1 << 30
			for _, p := range desc.Parameters {
				if p.Group == nil {
					// No advanced parameter is left alone in "Additional Settings" on platforms with groups.
					assert.False(t, p.Advanced != nil && *p.Advanced, "parameter %s", p.Name)
					lastOrderOutside = max(lastOrderOutside, *p.Order)
					continue
				}
				assert.Equal(t, groupTrafficFilter, *p.Group)
				// Platforms without groups keep showing the filters in "Additional Settings".
				assert.True(t, p.Advanced != nil && *p.Advanced, "parameter %s", p.Name)
				grouped = append(grouped, p.Name)
				firstOrderInside = min(firstOrderInside, *p.Order)
			}
			assert.Equal(t, []string{"headers", "headersMatchType", "sourceLabels"}, grouped)
			assert.Less(t, lastOrderOutside, firstOrderInside, "the filters come after the attack's own parameters")
			assert.Equal(t, []action_kit_api.ParameterGroup{trafficFilterGroup}, *desc.ParameterGroups)
		})
	}
}
