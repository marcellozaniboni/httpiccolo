@echo off
rem ## use this script to build a small executable for release purposes

set UPX_EXE="C:\utils\upx-3.96-win64\upx.exe"
if exist bin\httpiccolo.exe del bin\httpiccolo.exe

cd src
go build -o ..\bin\httpiccolo.exe -trimpath -ldflags="-s -w" .
cd ..

if exist "%UPX_EXE%" "%UPX_EXE%" --brute bin\httpiccolo.exe

dir bin\httpiccolo.exe