#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login

rm -rf .docs
mkdir -p .docs
./build/scheck-windows-amd64/scheck.exe -doc -dir .docs
/c/Users/18332/AppData/Roaming/npm/waque upload .docs/*.md
