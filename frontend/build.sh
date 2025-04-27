#!/bin/bash

echo "Building for windows"
GOOS=windows GOARCH=amd64 go build -o message-cli.exe .


GOOS=linux nGOARCH=amd64 go build -o message-cli .


