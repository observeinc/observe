#!/bin/bash
rsync -avh $HOME/observe/code/go/src/observe/cmd/observe/ $HOME/github.com/observeinc/observe/v1/
git add .
git commit
