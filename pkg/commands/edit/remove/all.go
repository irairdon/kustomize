/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package remove

import (
	"github.com/spf13/cobra"
	"github.com/irairdon/kustomize/v3/pkg/fs"
	"github.com/irairdon/kustomize/v3/pkg/ifc"
)

// NewCmdRemove returns an instance of 'remove' subcommand.
func NewCmdRemove(
	fsys fs.FileSystem,
	ldr ifc.Loader) *cobra.Command {
	c := &cobra.Command{
		Use:   "remove",
		Short: "Removes items from the kustomization file.",
		Long:  "",
		Example: `
	# Removes resources from the kustomization file
	kustomize edit remove resource {filepath} {filepath}
	kustomize edit remove resource {pattern}

	# Removes one or more patches from the kustomization file
	kustomize edit remove patch <filepath>

	# Removes one or more commonLabels from the kustomization file
	kustomize edit remove label {labelKey1},{labelKey2}

	# Removes one or more commonAnnotations from the kustomization file
	kustomize edit remove annotation {annotationKey1},{annotationKey2}
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdRemoveResource(fsys),
		newCmdRemoveLabel(fsys, ldr.Validator().MakeLabelNameValidator()),
		newCmdRemoveAnnotation(fsys, ldr.Validator().MakeAnnotationNameValidator()),
		newCmdRemovePatch(fsys),
	)
	return c
}
