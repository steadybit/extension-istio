// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extconfig

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"strings"
)

type Specification struct {
	ClusterName                              string   `required:"true" split_words:"true"`
	DiscoveryAttributesExcludesVirtualSerice []string `json:"discoveryAttributesExcludesVirtualSerice" split_words:"true" required:"false"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
}

func ValidateConfiguration() {
	// envconfig's `required:"true"` only checks that the variable is *set*: an empty
	// value satisfies it, so the extension would start with a blank configuration and
	// fail much later against the target system. Reject blank values here instead.
	if strings.TrimSpace(Config.ClusterName) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_CLUSTER_NAME must not be empty.")
	}
}
