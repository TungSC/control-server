#!/bin/bash
set -eux

APP_DIR=$(cd `dirname $0`/../; pwd)
cd $APP_DIR
mkdir -p $APP_DIR/logs
EXE=control-server
COMMAND=$APP_DIR/bin/$EXE

## build first
echo "go build -o control-server"
go build -o $COMMAND

$COMMAND
