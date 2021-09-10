#!/usr/bin/env bash

if [ "$1" = "bash" ]; then
  exec /bin/bash
fi

exec /usr/local/next-terminal/next-terminal "$@"