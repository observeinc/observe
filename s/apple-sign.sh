#!/bin/bash
set -x
codesign --verbose -s '8C3A8B6B52BD00B5F32E18ADFE95C9A8D6909711' --options runtime "$1"
xcrun notarytool submit "$1" --keychain-profile "notarytool-jwatte-gmail-observeinc-team" --wait
