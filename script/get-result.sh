#!/usr/bin/env bash

readonly help_usage=\
"usage: $0 [ID]
  ID: the id of the function call
"

if [ $# != 1 ]; then
  echo "$help_usage"
  exit 255
fi

id=$1

curl -X GET "$API_SERVER:$PORT/api/funcs/$id"
echo
