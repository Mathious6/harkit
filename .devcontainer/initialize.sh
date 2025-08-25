#!/bin/bash

if [ "$(uname)" == "Darwin" ]; then
    CHARLES_APP=$(mdfind "kMDItemCFBundleIdentifier == 'com.xk72.Charles'" | head -n 1)

    if [ -d "$CHARLES_APP" ]; then
        rm -rf .certs
        mkdir -p .certs
        "$CHARLES_APP/Contents/MacOS/Charles" ssl export .certs/charles-ssl.pem
    else
        echo "Charles is not installed, skipping certificate export."
    fi
fi
