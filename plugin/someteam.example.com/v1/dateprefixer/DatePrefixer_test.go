// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"github.com/irairdon/kustomize/v3/pkg/kusttest"
	plugins_test "github.com/irairdon/kustomize/v3/pkg/plugins/test"
)

func TestDatePrefixerPlugin(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	m := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: whatever
`,
		`apiVersion: apps/v1
kind: MeatBall
metadata:
  name: meatball
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: MeatBall
metadata:
  name: 2018-05-11-meatball
`)
}
