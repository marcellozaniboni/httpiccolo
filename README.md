# httpiccolo
a small http file server

## About httpiccolo

This project was initially born for personal needs: I just needed a minimal web file server, to be installed on my own server and then I started to code it using Go.

## Features

* Very easy to install: it's just a static binary with no dependencies.
* Very easy to configure: on first start the user is asked only four questions and the rest of the configuration is done using the administration web interface.
* All the configuration is saved in three json files within a directory reserved for this purpose. The configuration can be modified from the administration web interface (but the files can also be easily modified manually, if needed). You can reset the configuration pointing to another directory or by simply deleting it.

## Status

At the moment the software is still in beta, but it works, it's quite stable and hopefully without security bugs. So and I started using it for personal purposes on my servers.

The administration web interface can certainly be improved before the version 1.0 will be released. I also have already planned to add new features after 1.0.

## First run

1. Download the executable for your OS.
3. Just put the executable file in your filesystem and run it.
4. You will be asked for questions and the the configuration directory will be created.
5. By default a configuration subdirectory will be created in the same directory as the executable file. If you want to use another path, use the command line argument -c to specify it.
6. You can run "httpiccolo -h" to have a list of available command line arguments.

## How to build

Install Go (minimum version 1.19).

Quick compile and run:

* for Windows: run.bat
* for Linux: run.sh

To build an optimized release executable, install [UPX](https://upx.github.io) and

* for Windows: edit build-release.bat and fix UPX path, then run build-release.bat
* for Linux: run build-release.sh

## Compiled binaries

* Windows x86 64-bit [link]
* Linux x86 64-bit [link]
* Linux ARM 64-bit [link]
* Linux ARM 32-bit (armv6l) [link]
