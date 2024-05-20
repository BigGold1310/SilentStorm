#! /usr/bin/env sh

installOperatorSDK () {
    if ! type "operator-sdk" > /dev/null; then
        ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
        OS=$(uname | awk '{print tolower($0)}')
        OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.34.1
        curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
        mkdir -p /home/${USER}/bin
        mv operator-sdk_${OS}_${ARCH} /home/${USER}/bin/operator-sdk
        chmod +x /home/${USER}/bin/operator-sdk
    fi
}

cd /workspaces/silentstorm
go mod tidy
installOperatorSDK
