#!/bin/bash
#
# Generate the Go code for the generator and
# transformer factory functions in
#
#   github.com/irairdon/kustomize/v3/plugin/builtin
#
# from the raw plugin directories found _below_
# that directory.

set -e

myGoPath=$1
if [ -z ${1+x} ]; then
  myGoPath=$GOPATH
fi

if [ -z "$myGoPath" ]; then
  echo "Must specify a GOPATH"
  exit 1
fi

dir=$myGoPath/src/github.com/irairdon/kustomize

if [ ! -d "$dir" ]; then
  echo "$dir is not a directory."
  exit 1
fi

echo Generating linkable plugins...

pushd $dir >& /dev/null

GOPATH=$myGoPath go generate \
    github.com/irairdon/kustomize/v3/plugin/builtin/...
GOPATH=$myGoPath go fmt \
    github.com/irairdon/kustomize/v3/plugin/builtin

popd >& /dev/null

echo All done.
