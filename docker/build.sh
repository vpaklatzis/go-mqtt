#!/bin/sh

for dir in `ls -d */ | cut -f1 -d'/'`
do
    echo "Compiling $dir ...\c"
    cd $dir
    go clean
    GOOS=linux GOARCH=amd64 go build 
    cd ..
    echo " done."
done