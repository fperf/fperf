#!/bin/bash

files=$*

imports="auto_imports.go"
echo "package main" > $imports
function setup () {
    for f in $files;
    do
        importpath=`realpath $f | awk '{gsub(ENVIRON["GOPATH"]"/src/", "", $0); print}'`
        echo "import _ \"$importpath\"" >> $imports
    done
    touch fperf.go
}
function cleanup() {
    for f in $files;
    do
        rm -f $imports
    done
}

if [ $# = 0 ];then
    echo "Usage: $0 <testcase>"
    echo "For example"
    echo " Build mqtt testcase:"
    echo "  $0 testcases/mqtt"
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
