#!/usr/bin/env bash

version=${1}

[ -d "web" ] && (
pushd web
  cat package.json  | jq '.version="'${version}'"' > package.json.new
  mv package.json.new package.json
  [ ! -d "node_modules" ] && npm install
  npm run build
popd
)