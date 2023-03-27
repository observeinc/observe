#!/usr/bin/env bash
if [ -z "${AC_USERNAME}" ] || [ -z "${AC_PASSWORD}" ]; then
    echo "AC_USERNAME and AC_PASSWORD must be set."
    exit 1
fi
rm -rf dist
goreleaser build
