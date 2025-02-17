// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/irairdon/kustomize/v3/cmd/pluginator
package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/image"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/transformers"
	"github.com/irairdon/kustomize/v3/pkg/transformers/config"
	"sigs.k8s.io/yaml"
)

// Find matching image declarations and replace
// the name, tag and/or digest.
type plugin struct {
	ImageTag   image.Image        `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	FieldSpecs []config.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.ImageTag = image.Image{}
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		for _, path := range p.FieldSpecs {
			if !r.OrgId().IsSelected(&path.Gvk) {
				continue
			}
			err := transformers.MutateField(
				r.Map(), path.PathSlice(), false, p.mutateImage)
			if err != nil {
				return err
			}
		}
		// Kept for backward compatibility
		if err := p.findAndReplaceImage(r.Map()); err != nil && r.OrgId().Kind != `CustomResourceDefinition` {
			return err
		}
	}
	return nil
}

func (p *plugin) mutateImage(in interface{}) (interface{}, error) {
	original, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("image path is not of type string but %T", in)
	}

	if !isImageMatched(original, p.ImageTag.Name) {
		return original, nil
	}
	name, tag := split(original)
	if p.ImageTag.NewName != "" {
		name = p.ImageTag.NewName
	}
	if p.ImageTag.NewTag != "" {
		tag = ":" + p.ImageTag.NewTag
	}
	if p.ImageTag.Digest != "" {
		tag = "@" + p.ImageTag.Digest
	}
	return name + tag, nil
}

// findAndReplaceImage replaces the image name and
// tags inside one object.
// It searches the object for container session
// then loops though all images inside containers
// session, finds matched ones and update the
// image name and tag name
func (p *plugin) findAndReplaceImage(obj map[string]interface{}) error {
	paths := []string{"containers", "initContainers"}
	updated := false
	for _, path := range paths {
		containers, found := obj[path]
		if found {
			if _, err := p.updateContainers(containers); err != nil {
				return err
			}
			updated = true
		}
	}
	if !updated {
		return p.findContainers(obj)
	}
	return nil
}

func (p *plugin) updateContainers(in interface{}) (interface{}, error) {
	containers, ok := in.([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"containers path is not of type []interface{} but %T", in)
	}
	for i := range containers {
		container := containers[i].(map[string]interface{})
		containerImage, found := container["image"]
		if !found {
			continue
		}
		imageName := containerImage.(string)
		if isImageMatched(imageName, p.ImageTag.Name) {
			newImage, err := p.mutateImage(imageName)
			if err != nil {
				return nil, err
			}
			container["image"] = newImage
		}
	}
	return containers, nil
}

func (p *plugin) findContainers(obj map[string]interface{}) error {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			err := p.findAndReplaceImage(typedV)
			if err != nil {
				return err
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					err := p.findAndReplaceImage(typedItem)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func isImageMatched(s, t string) bool {
	// Tag values are limited to [a-zA-Z0-9_.-].
	pattern, _ := regexp.Compile("^" + t + "(@sha256)?(:[a-zA-Z0-9_.-]*)?$")
	return pattern.MatchString(s)
}

// split separates and returns the name and tag parts
// from the image string using either colon `:` or at `@` separators.
// Note that the returned tag keeps its separator.
func split(imageName string) (name string, tag string) {
	// check if image name contains a domain
	// if domain is present, ignore domain and check for `:`
	ic := -1
	if slashIndex := strings.Index(imageName, "/"); slashIndex < 0 {
		ic = strings.LastIndex(imageName, ":")
	} else {
		lastIc := strings.LastIndex(imageName[slashIndex:], ":")
		// set ic only if `:` is present
		if lastIc > 0 {
			ic = slashIndex + lastIc
		}
	}
	ia := strings.LastIndex(imageName, "@")
	if ic < 0 && ia < 0 {
		return imageName, ""
	}

	i := ic
	if ia > 0 {
		i = ia
	}

	name = imageName[:i]
	tag = imageName[i:]
	return
}
