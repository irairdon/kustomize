// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/irairdon/kustomize/v3/k8sdeps/kunstruct"
	"github.com/irairdon/kustomize/v3/k8sdeps/transformer"
	"github.com/irairdon/kustomize/v3/pkg/fs"
	kusttest_test "github.com/irairdon/kustomize/v3/pkg/kusttest"
	"github.com/irairdon/kustomize/v3/pkg/loader"
	"github.com/irairdon/kustomize/v3/pkg/plugins"
	plugins_test "github.com/irairdon/kustomize/v3/pkg/plugins/test"
	"github.com/irairdon/kustomize/v3/pkg/resmap"
	"github.com/irairdon/kustomize/v3/pkg/resource"
	"github.com/irairdon/kustomize/v3/pkg/target"
	"github.com/irairdon/kustomize/v3/pkg/validators"
)

func TestPluginDir(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "PrintWorkDir")

	base, err := os.Getwd()
	if err != nil {
		t.Fatalf("err %v", err)
	}
	dir, err := ioutil.TempDir(base, "kustomize-")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	defer os.RemoveAll(dir)

	fSys := fs.MakeRealFS()
	err = fSys.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte(`
generators:
- config.yaml
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}
	err = fSys.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
apiVersion: someteam.example.com/v1
kind: PrintWorkDir
metadata:
  name: some-random-name
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}

	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, validators.MakeFakeValidator(), dir, fSys)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	rf := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), nil)

	pl := plugins.NewLoader(plugins.ActivePluginConfig(), rf)
	tg, err := target.NewKustTarget(ldr, rf, transformer.NewFactoryImpl(), pl)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	m, err := tg.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	th := kusttest_test.NewKustTestHarness(t, ".")
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: WorkDir
metadata:
  name: `+dir+`
spec:
  path: `+dir+`
`)
}
