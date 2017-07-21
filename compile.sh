#!/bin/bash

go build -ldflags="-s -w"
upx dexm