#!/bin/bash

files=$*

function setup () {
    for f in $files;
    do
        name=`basename $f`
        cp $f client/$name
    done
}
function cleanup() {
    for f in $files;
    do
        name=`basename $f`
        rm -f client/$name
    done
}

if [ $# = 0 ];then
    echo "Usage: $0 <testcase>"
    echo "For example"
    echo " Build publish testcase:"
    echo "  $0 testcases/publish_client.go"
    echo " Or build all testcases:"
    echo "  $0 testcases/*"
    echo ""
    exit
fi

trap exit ERR SIGINT
setup
trap cleanup ERR SIGINT
make
cleanup
