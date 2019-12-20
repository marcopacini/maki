#!/usr/bin/env bash

FAILS=0

for file in **/*.maki; do
  OUTPUT=$(bash run_test.sh "./$file")
  if [ $? -eq 1 ]
  then
    FAILS=$((FAILS+1))
    echo "$file :: $OUTPUT"
  fi
done

if [ $FAILS -ne 0 ]
then
  exit 1
fi