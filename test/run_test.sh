#!/usr/bin/env bash

if [ $# -ne 1 ]
then
  echo 'usage: run_test.sh file'
  exit 1
fi

EXPECT=$(sed -n -e 's/^.*expect: //p' "$1")
OUTPUT=$(maki "$1" 2>&1)

if [ $? -eq 1 ]
then
  echo "FAILS: $OUTPUT"
  exit 1
fi

if [ "$OUTPUT" != "$EXPECT" ]
then
  echo "FAILS: output doesn't match"
  exit 1
fi