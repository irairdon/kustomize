// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/kr/pretty"

	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/resid"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/resource"
	"github.com/irairdon/kustomize/v3/pkg/transformers"
	"github.com/irairdon/kustomize/v3/pkg/types"
	"github.com/pkg/errors"
)

type Configurable interface {
	Config(ldr ifc.Loader, rf *resmap.Factory, config []byte) error
}

type Loader struct {
	pc *types.PluginConfig
	rf *resmap.Factory
}

func NewLoader(
	pc *types.PluginConfig, rf *resmap.Factory) *Loader {
	return &Loader{pc: pc, rf: rf}
}

func (l *Loader) LoadGenerators(
	ldr ifc.Loader, rm resmap.ResMap) ([]transformers.Generator, error) {
	var result []transformers.Generator
	for _, res := range rm.Resources() {
		g, err := l.LoadGenerator(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, nil
}

func (l *Loader) LoadGenerator(
	ldr ifc.Loader, res *resource.Resource) (transformers.Generator, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	g, ok := c.(transformers.Generator)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a generator", res.OrgId())
	}
	return g, nil
}

func (l *Loader) LoadTransformers(
	ldr ifc.Loader, rm resmap.ResMap) ([]transformers.Transformer, error) {
	var result []transformers.Transformer
	for _, res := range rm.Resources() {
		t, err := l.LoadTransformer(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (l *Loader) LoadTransformer(
	ldr ifc.Loader, res *resource.Resource) (transformers.Transformer, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	t, ok := c.(transformers.Transformer)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a transformer", res.OrgId())
	}
	return t, nil
}

func relativePluginPath(id resid.ResId) string {
	return filepath.Join(
		id.Group,
		id.Version,
		strings.ToLower(id.Kind))
}

func AbsolutePluginPath(pc *types.PluginConfig, id resid.ResId) string {
	return filepath.Join(
		pc.DirectoryPath, relativePluginPath(id), id.Kind)
}

func (l *Loader) absolutePluginPath(id resid.ResId) string {
	return AbsolutePluginPath(l.pc, id)
}

// TODO: https://github.com/kubernetes-sigs/kustomize/issues/1164
func (l *Loader) loadAndConfigurePlugin(
	ldr ifc.Loader, res *resource.Resource) (Configurable, error) {
	if !l.pc.Enabled {
		return nil, NotEnabledErr(res.OrgId().Kind)
	}
	c, err := l.loadPlugin(res.OrgId())
	if err != nil {
		return nil, err
	}
	yaml, err := res.AsYAML()
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling yaml from res %s", res.OrgId())
	}
	err = c.Config(ldr, l.rf, yaml)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s fails configuration", res.OrgId())
	}
	return c, nil
}

func (l *Loader) loadPlugin(resId resid.ResId) (Configurable, error) {
	p := NewExecPlugin(l.absolutePluginPath(resId))
	if p.isAvailable() {
		return p, nil
	}
	c, err := l.loadGoPlugin(resId)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// registry is a means to avoid trying to load the same .so file
// into memory more than once, which results in an error.
// Each test makes its own loader, and tries to load its own plugins,
// but the loaded .so files are in shared memory, so one will get
// "this plugin already loaded" errors if the registry is maintained
// as a Loader instance variable.  So make it a package variable.
var registry = make(map[string]Configurable)

func (l *Loader) loadGoPlugin(id resid.ResId) (Configurable, error) {
	regId := relativePluginPath(id)
	if c, ok := registry[regId]; ok {
		return copyPlugin(c), nil
	}
	absPath := l.absolutePluginPath(id)
	p, err := plugin.Open(absPath + ".so")
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails to load", absPath)
	}
	symbol, err := p.Lookup(PluginSymbol)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s doesn't have symbol %s",
			regId, PluginSymbol)
	}
	pretty.Println(symbol)
	c, ok := symbol.(Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", regId)
	}
	registry[regId] = c
	return copyPlugin(c), nil
}

func copyPlugin(c Configurable) Configurable {
	indirect := reflect.Indirect(reflect.ValueOf(c))
	newIndirect := reflect.New(indirect.Type())
	newIndirect.Elem().Set(reflect.ValueOf(indirect.Interface()))
	newNamed := newIndirect.Interface()
	return newNamed.(Configurable)
}
