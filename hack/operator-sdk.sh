#!/bin/bash

function download() {
    echo "downloading operator-sdk"
    export ARCH=$(case $(arch) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(arch) ;; esac)
    export OS=$(uname | awk '{print tolower($0)}')
    export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_${OS}_${ARCH}
    echo "url: $OPERATOR_SDK_DL_URL"
    curl -LO ${OPERATOR_SDK_DL_URL}
    chmod +x operator-sdk_${OS}_${ARCH}
    mkdir -p bin
    mv operator-sdk_${OS}_${ARCH} ./bin/operator-sdk
}

OPERATOR_SDK=$(ls ./bin/operator-sdk)
if [[ $? -eq 0 ]]; then
    echo "checking operator-sdk at ${OPERATOR_SDK}"
    CUR_VERSION=$(${OPERATOR_SDK} version 2>/dev/null)
    CUR_VERSION=$(echo ${CUR_VERSION} | sed 's/^.*version: \"\(v[^\"]*\).*$/\1/')
    if [[ "$CUR_VERSION" == "$OPERATOR_SDK_VERSION" ]]; then
        echo "correct operator-sdk found: ${CUR_VERSION}"
    else
        echo "wrong operator-sdk version: ${CUR_VERSION}"
        download
    fi
else
    echo "no operator-sdk found"
    download
fi
