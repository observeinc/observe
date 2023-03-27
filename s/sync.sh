#!/bin/bash
rm -f *.*
git restore go.mod go.sum
rsync -avh $HOME/observe/code/go/src/observe/cmd/observe/ $HOME/github.com/observeinc/observe/
git add .
git commit
