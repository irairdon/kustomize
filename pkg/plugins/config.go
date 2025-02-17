// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"fmt"
	"github.com/spf13/pflag"
	"path/filepath"
	"github.com/irairdon/kustomize/v3/pkg/pgmconfig"
	"github.com/irairdon/kustomize/v3/pkg/types"
)

const (
	PluginSymbol          = "KustomizePlugin"
	flagEnablePluginsName = "enable_alpha_plugins"
	flagEnablePluginsHelp = `enable plugins, an alpha feature.
See https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins.md
`
	flagErrorFmt = `
unable to load plugin %s because plugins disabled
specify the flag
  --%s
to %s`
)

func ActivePluginConfig() *types.PluginConfig {
	pc := DefaultPluginConfig()
	pc.Enabled = true
	return pc
}

func DefaultPluginConfig() *types.PluginConfig {
	return &types.PluginConfig{
		Enabled: false,
		DirectoryPath: filepath.Join(
			pgmconfig.ConfigRoot(), pgmconfig.PluginRoot),
	}
}

func NotEnabledErr(name string) error {
	return fmt.Errorf(
		flagErrorFmt,
		name,
		flagEnablePluginsName,
		flagEnablePluginsHelp)
}

func AddFlagEnablePlugins(set *pflag.FlagSet, v *bool) {
	set.BoolVar(
		v, flagEnablePluginsName,
		false, flagEnablePluginsHelp)
}
