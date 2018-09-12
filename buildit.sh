#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o ./mail-surfer-linux .
upx mail-surfer-linux

