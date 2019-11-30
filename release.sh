#!/bin/sh
go generate
GOOS=windows GOARCH=386 go build
GOOS=darwin GOARCH=amd64 go build
mv tiktak tiktak.mac
go build
. ./VERSION
git_tag=v$major.$minor.$patch-$quality+$build_no
echo $git_tag
