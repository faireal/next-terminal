#!/usr/bin/env bash

version=${1}

[ -d "web" ] && (
pushd web
  sed -nE -i 's/(^\s*"version": ")(.*?)(",$)/\${version}\3/p' package.json
  [ ! -d "node_modules" ] && npm install
  npm run build
popd
)