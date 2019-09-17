// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapandsecret

import (
	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/types"
)

// Factory makes ConfigMaps and Secrets.
type Factory struct {
	ldr     ifc.Loader
	options *types.GeneratorOptions
}

// NewFactory returns a new factory that makes ConfigMaps and Secrets.
func NewFactory(
	ldr ifc.Loader, o *types.GeneratorOptions) *Factory {
	return &Factory{ldr: ldr, options: o}
}

const keyExistsErrorMsg = "cannot add key %s, another key by that name already exists: %v"
