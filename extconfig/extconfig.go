// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extconfig

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Specification struct {
	ClusterName                              string `required:"true" split_words:"true"`
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
	// You may optionally validate the configuration here.
}
