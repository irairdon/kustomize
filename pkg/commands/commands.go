// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/irairdon/kustomize/v3/k8sdeps/kunstruct"
	"github.com/irairdon/kustomize/v3/k8sdeps/transformer"
	"github.com/irairdon/kustomize/v3/k8sdeps/validator"
	"github.com/irairdon/kustomize/v3/pkg/commands/build"
	"github.com/irairdon/kustomize/v3/pkg/commands/create"
	"github.com/irairdon/kustomize/v3/pkg/commands/edit"
	"github.com/irairdon/kustomize/v3/pkg/commands/misc"
	"github.com/irairdon/kustomize/v3/pkg/fs"
	"github.com/irairdon/kustomize/v3/pkg/pgmconfig"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/resource"
)

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand() *cobra.Command {
	fSys := fs.MakeRealFS()
	stdOut := os.Stdout

	c := &cobra.Command{
		Use:   pgmconfig.ProgramName,
		Short: "Manages declarative configuration of Kubernetes",
		Long: `
Manages declarative configuration of Kubernetes.
See https://github.com/irairdon/kustomize
`,
	}

	uf := kunstruct.NewKunstructuredFactoryImpl()
	pf := transformer.NewFactoryImpl()
	rf := resmap.NewFactory(resource.NewFactory(uf), pf)
	v := validator.NewKustValidator()
	c.AddCommand(
		build.NewCmdBuild(
			stdOut, fSys, v,
			rf, pf),
		edit.NewCmdEdit(fSys, v, uf),
		create.NewCmdCreate(fSys, uf),
		misc.NewCmdConfig(fSys),
		misc.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
