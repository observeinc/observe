#!/usr/bin/env bash
set -e
observe="$(command -v observe)"
_observe_completions()
{
    # The proposed work-arounds don't handle errors well, and the output will
    # never contain spaces anyway.
    # shellcheck disable=SC2207
    result=($(COMP_LINE="${COMP_LINE}" COMP_POINT="${COMP_POINT}" "${observe}" complete --shell=bash))
    COMPREPLY=("${result[@]}")
}
complete -F _observe_completions observe
