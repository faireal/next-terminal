#!/usr/bin/env bash

[ -d "web" ] && (
pushd web
  [ ! -d "node_modules" ] && npm install
  npm run build
popd
)