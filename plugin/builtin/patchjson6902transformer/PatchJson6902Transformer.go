// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/irairdon/kustomize/v3/cmd/pluginator
package main

import (
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"github.com/irairdon/kustomize/v3/pkg/gvk"
	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/resid"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	ldr          ifc.Loader
	decodedPatch jsonpatch.Patch
	Target       types.PatchTarget `json:"target,omitempty" yaml:"target,omitempty"`
	Path         string            `json:"path,omitempty" yaml:"path,omitempty"`
	JsonOp       string            `json:"jsonOp,omitempty" yaml:"jsonOp,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.ldr = ldr
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	if p.Target.Name == "" {
		return fmt.Errorf("must specify the target name")
	}
	if p.Path == "" && p.JsonOp == "" {
		return fmt.Errorf("empty file path and empty jsonOp")
	}
	if p.Path != "" {
		if p.JsonOp != "" {
			return fmt.Errorf("must specify a file path or jsonOp, not both")
		}
		rawOp, err := p.ldr.Load(p.Path)
		if err != nil {
			return err
		}
		p.JsonOp = string(rawOp)
		if p.JsonOp == "" {
			return fmt.Errorf("patch file '%s' empty seems to be empty", p.Path)
		}
	}
	if p.JsonOp[0] != '[' {
		// if it doesn't seem to be JSON, imagine
		// it is YAML, and convert to JSON.
		op, err := yaml.YAMLToJSON([]byte(p.JsonOp))
		if err != nil {
			return err
		}
		p.JsonOp = string(op)
	}
	p.decodedPatch, err = jsonpatch.DecodePatch([]byte(p.JsonOp))
	if err != nil {
		return errors.Wrapf(err, "decoding %s", p.JsonOp)
	}
	if len(p.decodedPatch) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; file=%s, JsonOp=%s", p.Path, p.JsonOp)
	}
	return err
}

func (p *plugin) Transform(m resmap.ResMap) error {
	id := resid.NewResIdWithNamespace(
		gvk.Gvk{
			Group:   p.Target.Group,
			Version: p.Target.Version,
			Kind:    p.Target.Kind,
		},
		p.Target.Name,
		p.Target.Namespace,
	)
	obj, err := m.GetById(id)
	if err != nil {
		return err
	}
	rawObj, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	modifiedObj, err := p.decodedPatch.Apply(rawObj)
	if err != nil {
		return errors.Wrapf(
			err, "failed to apply json patch '%s'", p.JsonOp)
	}
	return obj.UnmarshalJSON(modifiedObj)
}
