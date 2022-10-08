@echo off
rem ## use this script to compile and run the application

if exist bin\httpiccolo.exe del bin\httpiccolo.exe

cd src
go build -o ..\bin\httpiccolo.exe .
cd ..
dir bin\httpiccolo.exe

bin\httpiccolo.exe
