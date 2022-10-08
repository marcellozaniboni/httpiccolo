#!/bin/bash
#
# use this script to compile and run the application

rm -f ./bin/httpiccolo
cd src
go build -o ../bin/httpiccolo .
cd ..

ls -lph ./bin/httpiccolo
./bin/httpiccolo
