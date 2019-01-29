#!/usr/bin/env bash

APP_NAME=$1

rm -rf app/*
rm -rf /tmp/build

# Build go app
echo "Moving source code"
# Make build folder
mkdir -p /tmp/build/$APP_NAME
cp -r ../* /tmp/build/$APP_NAME

docker_cmds="echo Begin building go binary; \

cd /tmp/build/$APP_NAME; \

echo running go build; \
go build cmd/server/main.go; \
echo Done building go binary"

docker run --rm -v /tmp/build/:/tmp/build -w /tmp/build -e GO111MODULE=on -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.11.1 /bin/bash -c "$docker_cmds"

mkdir -p app/
cp /tmp/build/$APP_NAME/main app/main