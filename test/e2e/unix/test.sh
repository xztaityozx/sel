#!/bin/bash

set -e

SEL=../../dist/sel

ls -1 | grep -v test.sh | \
  while read DIR; do
    commandline=$(cat ./$DIR/commandline)
    input=$DIR/input
    output=$DIR/output
    echo -en "test $DIR: $SEL $commandline ... "
    cat $input | eval "$SEL $commandline" | diff - $output && echo "OK" || ( echo "NG"; false )
  done
