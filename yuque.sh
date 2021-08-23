#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login

rm -rf .docs
mkdir -p .docs
./build/scheck-darwin-amd64/scheck -doc -dir .docs
#waque upload .docs/*.md
