// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/irairdon/kustomize/v3/cmd/pluginator
package main

import (
	"fmt"

	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/resid"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/resource"
	"github.com/irairdon/kustomize/v3/pkg/transformers"
	"github.com/irairdon/kustomize/v3/pkg/transformers/config"
	"github.com/irairdon/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

// Change or set the namespace of non-cluster level resources.
type plugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	FieldSpecs       []config.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.Namespace = ""
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Namespace) == 0 {
		return nil
	}
	for _, r := range m.Resources() {
		if len(r.Map()) == 0 {
			// Don't mutate empty objects?
			continue
		}

		id := r.OrgId()
		applicableFs := p.applicableFieldSpecs(id)

		for _, fs := range applicableFs {
			err := transformers.MutateField(
				r.Map(), fs.PathSlice(), fs.CreateIfNotPresent,
				p.changeNamespace(r))
			if err != nil {
				return err
			}
		}

		matches := m.GetMatchingResourcesByCurrentId(r.CurId().Equals)
		if len(matches) != 1 {
			return fmt.Errorf("namespace tranformation produces ID conflict: %#v", matches)
		}
	}
	return nil
}

const metaNamespace = "metadata/namespace"

// Special casing metadata.namespace since
// all objects have it, even "ClusterKind" objects
// that don't exist in a namespace (the Namespace
// object itself doesn't live in a namespace).
func (p *plugin) applicableFieldSpecs(id resid.ResId) []config.FieldSpec {
	res := []config.FieldSpec{}
	for _, fs := range p.FieldSpecs {
		if id.IsSelected(&fs.Gvk) && (fs.Path != metaNamespace || (fs.Path == metaNamespace && id.IsNamespaceableKind())) {
			res = append(res, fs)
		}
	}
	return res
}

func (o *plugin) changeNamespace(
	referrer *resource.Resource) func(in interface{}) (interface{}, error) {
	return func(in interface{}) (interface{}, error) {
		switch in.(type) {
		case string:
			// will happen when the metadata/namespace
			// value is replaced
			return o.Namespace, nil
		case []interface{}:
			l, _ := in.([]interface{})
			for idx, item := range l {
				switch item.(type) {
				case map[string]interface{}:
					// Will happen when mutating the subjects
					// field of ClusterRoleBinding and RoleBinding
					inMap, _ := item.(map[string]interface{})
					if _, ok := inMap["name"]; !ok {
						continue
					}
					name, ok := inMap["name"].(string)
					if !ok {
						continue
					}
					// The only case we need to force the namespace
					// if for the "service account". "default" is
					// kind of hardcoded here for right now.
					if name != "default" {
						continue
					}
					inMap["namespace"] = o.Namespace
					l[idx] = inMap
				default:
					// nothing to do for right now
				}
			}
			return in, nil
		case map[string]interface{}:
			// Will happen if the createField=true
			// when the namespace is added to the
			// object
			inMap := in.(map[string]interface{})
			if len(inMap) == 0 {
				return o.Namespace, nil
			} else {
				return in, nil
			}
		default:
			return in, nil
		}
	}
}
