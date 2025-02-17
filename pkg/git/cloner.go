/*
Copyright 2018 The Kubernetes Authors.

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

package git

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/irairdon/kustomize/v3/pkg/fs"
)

// Cloner is a function that can clone a git repo.
type Cloner func(repoSpec *RepoSpec) error

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec) error {
	gitProgram, err := exec.LookPath("git")
	if err != nil {
		return errors.Wrap(err, "no 'git' program on path")
	}
	repoSpec.Dir, err = fs.NewTmpConfirmedDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(
		gitProgram,
		"init",
		repoSpec.Dir.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		log.Printf("Error initializing empty git repo: %s", out.String())
		return errors.Wrapf(
			err,
			"trouble initializing empty git repo in %s",
			repoSpec.Dir.String())
	}

	cmd = exec.Command(
		gitProgram,
		"remote",
		"add",
		"origin",
		repoSpec.CloneSpec())
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = repoSpec.Dir.String()
	err = cmd.Run()
	if err != nil {
		log.Printf("Error setting git remote: %s", out.String())
		return errors.Wrapf(
			err,
			"trouble adding remote %s",
			repoSpec.CloneSpec())
	}
	if repoSpec.Ref == "" {
		repoSpec.Ref = "master"
	}
	cmd = exec.Command(
		gitProgram,
		"fetch",
		"--depth=1",
		"origin",
		repoSpec.Ref)
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = repoSpec.Dir.String()
	err = cmd.Run()
	if err != nil {
		log.Printf("Error performing git fetch: %s", out.String())
		return errors.Wrapf(err, "trouble fetching %s", repoSpec.Ref)
	}

	cmd = exec.Command(
		gitProgram,
		"reset",
		"--hard",
		"FETCH_HEAD")
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = repoSpec.Dir.String()
	err = cmd.Run()
	if err != nil {
		log.Printf("Error performing git reset: %s", out.String())
		return errors.Wrapf(
			err, "trouble hard resetting empty repository to %s", repoSpec.Ref)
	}

	cmd = exec.Command(
		gitProgram,
		"submodule",
		"update",
		"--init",
		"--recursive")
	cmd.Stdout = &out
	cmd.Dir = repoSpec.Dir.String()
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "trouble fetching submodules for %s", repoSpec.Ref)
	}

	return nil
}

// DoNothingCloner returns a cloner that only sets
// cloneDir field in the repoSpec.  It's assumed that
// the cloneDir is associated with some fake filesystem
// used in a test.
func DoNothingCloner(dir fs.ConfirmedDir) Cloner {
	return func(rs *RepoSpec) error {
		rs.Dir = dir
		return nil
	}
}
