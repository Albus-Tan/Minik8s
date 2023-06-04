#!/usr/bin/env bash

readonly help_usage="usage: $0 [FUNCTION_NAME] [ARG]
  CONFIG_NAME: the name of the function defined previously
  ARG: string of the function argument
"

if [ $# != 2 ]; then
  echo "$help_usage"
  exit 255
fi

func=$1
arg=$2

curl -X POST "$API_SERVER:$PORT/api/funcs/$func" --data "$arg"
echo
