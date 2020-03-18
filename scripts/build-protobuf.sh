#!/usr/bin/env sh

error_message="Couldn't find any .proto files.\nThis script should be run in project root."
if ! ls ./blockchain/*.proto 1>/dev/null 2>&1; then
    echo $error_message
    exit 1
fi
if ! ls ./network/protocol/*.proto 1>/dev/null 2>&1; then
    echo $error_message
    exit 1
fi

protoc -I=. --go_out=. ./blockchain/*.proto || exit 1
protoc -I=. --go_out=. ./network/protocol/*.proto || exit 1
