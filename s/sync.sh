#!/bin/bash
rsync -avh $HOME/observe/code/go/src/observe/cmd/observe/ $HOME/github.com/observeinc/observe/
git add .
git commit
