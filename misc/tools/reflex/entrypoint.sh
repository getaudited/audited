#!/bin/bash

cd /usr/app
go run ./misc/tools/reflex &&
task run:$APP_NAME
