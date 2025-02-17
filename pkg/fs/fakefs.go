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

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/irairdon/kustomize/v3/pkg/pgmconfig"
)

var _ FileSystem = &fakeFs{}

// fakeFs implements FileSystem using a fake in-memory filesystem.
type fakeFs struct {
	m map[string]*FakeFile
}

// MakeFakeFS returns an instance of fakeFs with no files in it.
func MakeFakeFS() *fakeFs {
	result := &fakeFs{m: map[string]*FakeFile{}}
	result.Mkdir("/")
	return result
}

// kustomizationContent is used in tests.
const kustomizationContent = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: some-prefix
nameSuffix: some-suffix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
commonLabels:
  app: helloworld
commonAnnotations:
  note: This is an example annotation
resources: []
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
configMapGenerator: []
# There could be secrets in Base, if just using a fork/rebase workflow
secretGenerator: []
`

// Create assures a fake file appears in the in-memory file system.
func (fs *fakeFs) Create(name string) (File, error) {
	f := &FakeFile{}
	f.open = true
	fs.m[name] = f
	return fs.m[name], nil
}

// Mkdir assures a fake directory appears in the in-memory file system.
func (fs *fakeFs) Mkdir(name string) error {
	fs.m[name] = makeDir(name)
	return nil
}

// MkdirAll delegates to Mkdir
func (fs *fakeFs) MkdirAll(name string) error {
	return fs.Mkdir(name)
}

// RemoveAll presumably does rm -r on a path.
// There's no error.
func (fs *fakeFs) RemoveAll(name string) error {
	var toRemove []string
	for k := range fs.m {
		if strings.HasPrefix(k, name) {
			toRemove = append(toRemove, k)
		}
	}
	for _, k := range toRemove {
		delete(fs.m, k)
	}
	return nil
}

// Open returns a fake file in the open state.
func (fs *fakeFs) Open(name string) (File, error) {
	if _, found := fs.m[name]; !found {
		return nil, fmt.Errorf("file %q cannot be opened", name)
	}
	return fs.m[name], nil
}

// CleanedAbs cannot fail.
func (fs *fakeFs) CleanedAbs(path string) (ConfirmedDir, string, error) {
	if fs.IsDir(path) {
		return ConfirmedDir(path), "", nil
	}
	d := filepath.Dir(path)
	if d == path {
		return ConfirmedDir(d), "", nil
	}
	return ConfirmedDir(d), filepath.Base(path), nil
}

// Exists returns true if file is known.
func (fs *fakeFs) Exists(name string) bool {
	_, found := fs.m[name]
	return found
}

// Glob returns the list of matching files
func (fs *fakeFs) Glob(pattern string) ([]string, error) {
	var result []string
	for p := range fs.m {
		if fs.pathMatch(p, pattern) {
			result = append(result, p)
		}
	}
	sort.Strings(result)
	return result, nil
}

// IsDir returns true if the file exists and is a directory.
func (fs *fakeFs) IsDir(name string) bool {
	f, found := fs.m[name]
	if found && f.dir {
		return true
	}
	if !strings.HasSuffix(name, "/") {
		name = name + "/"
	}
	for k := range fs.m {
		if strings.HasPrefix(k, name) {
			return true
		}
	}
	return false
}

// ReadFile always returns an empty bytes and error depending on content of m.
func (fs *fakeFs) ReadFile(name string) ([]byte, error) {
	if ff, found := fs.m[name]; found {
		return ff.content, nil
	}
	return nil, fmt.Errorf("cannot read file %q", name)
}

func (fs *fakeFs) ReadTestKustomization() ([]byte, error) {
	return fs.ReadFile(pgmconfig.KustomizationFileNames[0])
}

// WriteFile always succeeds and does nothing.
func (fs *fakeFs) WriteFile(name string, c []byte) error {
	ff := &FakeFile{}
	ff.Write(c)
	fs.m[name] = ff
	return nil
}

// WriteTestKustomization writes a standard test file.
func (fs *fakeFs) WriteTestKustomization() {
	fs.WriteTestKustomizationWith([]byte(kustomizationContent))
}

// WriteTestKustomizationWith writes a standard test file.
func (fs *fakeFs) WriteTestKustomizationWith(bytes []byte) {
	fs.WriteFile(pgmconfig.KustomizationFileNames[0], bytes)
}

// Walk implements filepath.Walk using the fake filesystem.
func (fs *fakeFs) Walk(path string, walkFn filepath.WalkFunc) error {
	info, err := fs.lstat(path)
	if err != nil {
		err = walkFn(path, info, err)
	} else {
		err = fs.walk(path, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func (fs *fakeFs) pathMatch(path, pattern string) bool {
	match, _ := filepath.Match(pattern, path)
	return match
}

func (fs *fakeFs) lstat(path string) (*Fakefileinfo, error) {
	f, found := fs.m[path]
	if !found {
		return nil, os.ErrNotExist
	}
	return &Fakefileinfo{f}, nil
}

func (fs *fakeFs) join(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return strings.Replace(strings.Join(elem[i:], "/"), "//", "/", -1)
		}
	}
	return ""
}

func (fs *fakeFs) readDirNames(path string) []string {
	var names []string
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	pathSegments := strings.Count(path, "/")
	for name := range fs.m {
		if name == path {
			continue
		}
		if strings.Count(name, "/") > pathSegments {
			continue
		}
		if strings.HasPrefix(name, path) {
			names = append(names, filepath.Base(name))
		}
	}
	sort.Strings(names)
	return names
}

func (fs *fakeFs) walk(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names := fs.readDirNames(path)
	if err := walkFn(path, info, nil); err != nil {
		return err
	}
	for _, name := range names {
		filename := fs.join(path, name)
		fileInfo, err := fs.lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, os.ErrNotExist); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = fs.walk(filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}
