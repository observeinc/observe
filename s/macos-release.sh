#!/usr/bin/env bash
if [ -z "${AC_USERNAME}" ] || [ -z "${AC_PASSWORD}" ]; then
    echo "AC_USERNAME, AC_PASSWORD, AC_OUTPUT_DMG and AC_OUTPUT_ZIP must be set."
    exit 1
fi
if [ ! -x ./observe ]; then
    echo "Please build with s/build.sh first."
    exit 1
fi
