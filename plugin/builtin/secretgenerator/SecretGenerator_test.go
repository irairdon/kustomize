// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"github.com/irairdon/kustomize/v3/pkg/kusttest"
	plugins_test "github.com/irairdon/kustomize/v3/pkg/plugins/test"
)

func TestSecretGenerator(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "SecretGenerator")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/a.env", `
ROUTER_PASSWORD=admin
`)
	th.WriteF("/app/b.env", `
DB_PASSWORD=iloveyou
`)
	th.WriteF("/app/longsecret.txt", `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit.
`)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: mySecret
  namespace: whatever
behavior: merge
envs:
- a.env
- b.env
files:
- obscure=longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  DB_PASSWORD: aWxvdmV5b3U=
  FRUIT: YXBwbGU=
  ROUTER_PASSWORD: YWRtaW4=
  VEGETABLE: Y2Fycm90
  obscure: CkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LApjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIGVsaXQuCg==
kind: Secret
metadata:
  name: mySecret
  namespace: whatever
type: Opaque
`)
}
