#!/bin/bash
readarray -t lines < $1
username=${lines[0]}
password=${lines[1]}
# Replace your own authentication mechanism here
if [[ "$password" == "bao" ]]; then
  echo "ok"
  exit 0
fi
echo "not ok"
exit 1