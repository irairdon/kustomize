// Code generated by pluginator on HashTransformer; DO NOT EDIT.
package builtin

import (
	"fmt"

	"github.com/irairdon/kustomize/v3/pkg/ifc"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
)

type HashTransformerPlugin struct {
	hasher ifc.KunstructuredHasher
}

//noinspection GoUnusedGlobalVariable
func NewHashTransformerPlugin() *HashTransformerPlugin {
	return &HashTransformerPlugin{}
}

func (p *HashTransformerPlugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, config []byte) (err error) {
	p.hasher = rf.RF().Hasher()
	return nil
}

// Transform appends hash to generated resources.
func (p *HashTransformerPlugin) Transform(m resmap.ResMap) error {
	for _, res := range m.Resources() {
		if res.NeedHashSuffix() {
			h, err := p.hasher.Hash(res)
			if err != nil {
				return err
			}
			res.SetName(fmt.Sprintf("%s-%s", res.GetName(), h))
		}
	}
	return nil
}
