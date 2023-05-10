#!/bin/bash
set -x
osslsigncode sign -spc "$AUTHENTICODE_SPC" -key "$AUTHENTICODE_KEY" -t http://timestamp.digicert.com -in "$1" -out "$1.out"
mv "$1.out" "$1"
