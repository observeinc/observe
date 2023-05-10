#!/bin/bash
if [ -z "$1" ]; then
    echo "usage: s/apple-sign.sh artifact-path.zip"
    exit 1
fi
set -x
xcrun notarytool submit "$1" --keychain-profile "notarytool-jwatte-gmail-observeinc-team" --wait
