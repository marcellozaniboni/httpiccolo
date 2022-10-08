#!/bin/bash
#
# use this script to build a small executable for release purposes

t1=$(date +%s)
rm -f ./bin/httpiccolo
cd src
go build -o ../bin/httpiccolo -trimpath -ldflags="-s -w" .
cd ..
if [ "$(which upx)" != "" ]; then
    upx --brute ./bin/httpiccolo
fi
t2=$(date +%s)

ls -lh ./bin/httpiccolo

tot_t=$((t2 - t1))
echo "total build time: $tot_t seconds"
