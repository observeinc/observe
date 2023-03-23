#!/bin/bash
set -xeu -o pipefail
cd "$HOME/github.com/observeinc/observe/v2"
RELVER="$(cat ../release.txt)"
RELNAME="observeinc-observe-${RELVER}-$(uname -s | tr '[A-Z]' '[a-z]')-$(uname -m)"
DIRNAME="../release/$RELNAME"
go build -ldflags "-X main.GitCommit=$(git rev-parse HEAD) -X main.ReleaseName=${RELNAME}" .
rm -rf "${DIRNAME}"
mkdir -p "${DIRNAME}"
mv observe "${DIRNAME}/"
cp README.md "${DIRNAME}/"
cd "../release"
zip -r "${RELNAME}.zip" "${RELNAME}"
