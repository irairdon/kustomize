// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/irairdon/kustomize/v3/cmd/pluginator
package main

import (
	"github.com/pkg/errors"
	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/resource"
	"sort"
)

// Sort the resources using an ordering defined in the Gvk class.
// This puts cluster-wide basic resources with no
// dependencies (like Namespace, StorageClass, etc.)
// first, and resources with a high number of dependencies
// (like ValidatingWebhookConfiguration) last.
type plugin struct{}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	resources := make([]*resource.Resource, m.Size())
	ids := m.AllIds()
	sort.Sort(resmap.IdSlice(ids))
	for i, id := range ids {
		resources[i], err = m.GetByCurrentId(id)
		if err != nil {
			return errors.Wrap(err, "expected match for sorting")
		}
	}
	m.Clear()
	for _, r := range resources {
		m.Append(r)
	}
	return nil
}
