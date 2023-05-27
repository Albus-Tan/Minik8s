#!/usr/bin/env bash
readonly DEFAULT_VERSION="v1"

readonly ARGS=(
  "API_SERVER"
  "PORT"
  "NAME"
  "ADDR"
  "PRE_RUN"
  "MAIN"
  "LEFT_BRANCH"
  "RIGHT_BRANCH"
)

readonly help_usage=\
"usage: $0 [CONFIG_FILE]
  CONFIG_FILE: a file contains lines of [NAME]=[VALUE]
  NAME: should be one of these: ${ARGS[*]} VERSION
  VALUE: string"

readonly help_config_missing=\
"config file is missing:"

readonly help_var_undefine=\
"var is missing:"

readonly help_version_undefine=\
"VERSION is missing, use default: $DEFAULT_VERSION"

readonly help_file_missing=\
"file is missing: "

if [ $# != 1 ]; then
  echo  "$help_usage"
  exit 255
fi

conf="$1"

if [ ! -e "$conf" ]; then
  echo "$help_config_missing $conf"
  exit 255
fi


source $conf



for arg in "${ARGS[@]}" ; do
  if [ ! -v "$arg" ]; then
    echo "$help_var_undefine $arg"
    exit 255
  fi
done

if [ ! -v VERSION ]; then
  echo "$help_version_undefine"
  VERSION=$DEFAULT_VERSION
fi

readonly FILES=(
  "$PRE_RUN"
  "$MAIN"
)


for file in "${FILES[@]}" ; do
  if [ ! -e "$file" ]; then
    echo "$help_file_missing $file"
    exit 255
  fi
done

data=\
"{
  \"apiVersion\": \"$VERSION\",
  \"kind\": \"Func\",
  \"metadata\": {
    \"name\": \"$NAME\",
    \"namespace\": \"default\"
  },
  \"spec\": {
    \"name\":\"$NAME\",
    \"preRun\":\"$(base64  "$PRE_RUN"| tr -d '\n')\",
    \"function\":\"$(base64  "$MAIN"| tr -d '\n')\",
    \"left\":\"$LEFT_BRANCH\",
    \"right\":\"$RIGHT_BRANCH\",
    \"serviceAddress\":\"$ADDR\"
  }
}"

curl -X POST "$API_SERVER:8080/api/funcs/template/" --data "$data"
echo
