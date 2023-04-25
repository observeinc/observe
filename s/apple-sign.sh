#!/bin/bash
set -x
xcrun notarytool submit "$1" --keychain-profile "notarytool-jwatte-gmail-observeinc-team" --wait
