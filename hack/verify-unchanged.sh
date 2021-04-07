#!/bin/bash

if [[ -n "$(git status --porcelain .)" ]]; then
    echo "Uncommitted files. Run 'make lint' and commit modified files."
    echo "$(git status --porcelain .)"
    exit 1
fi
