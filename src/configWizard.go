package main

// httpiccolo
//
// Copyright © 2022 Marcello Zaniboni
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public
// License along with this program; if not, you can visit
// https://www.gnu.org/licenses/
// or write to Free Software Foundation, Inc.,
// 675 Mass Ave, Cambridge, MA 02139, USA.

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
	"marcellozaniboni.net/httpiccolo/mdao"
	"marcellozaniboni.net/httpiccolo/mutils"
)

func configWizard(directory string) {

	fmt.Println("NEW CONFIGURATION")
	fmt.Println("Please answer four questions.")

	// admin username
	fmt.Print("\nAdministrator username (e.g. many people use \"admin\")? ")
	adminUsername := mutils.ReadStdinLine()
	if adminUsername == "" {
		mutils.FatalMessage("Error: the username cannot be empty.")
	}

	// admin password
	fmt.Print("Administrator password (at least 5 characters long)? ")
	bytepassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		mutils.FatalError("fatal error", err)
	}
	adminPassword := string(bytepassword)
	fmt.Println()
	if len(adminPassword) < 5 {
		mutils.FatalMessage("Error: the password must be at least 5 characters long.")
	}

	// server port
	fmt.Print("HTTP port number (e.g. many people use 8080 or 80)? ")
	stringport := mutils.ReadStdinLine()
	port, err := strconv.Atoi(stringport)
	if err != nil {
		mutils.FatalMessage("Error: insert a number.")
	}
	if port < 20 || port > 65535 {
		mutils.FatalMessage("Error: insert a number between 20 and 65535.")
	}

	// root directory
	fmt.Println("\nNow enter the root directory for all the published contents.")
	fmt.Println("Please avoid relative path like ./www")
	fmt.Println("Example for absolute an path on Windows:\n\tC:/webtools/httpiccolo/webcontent\n\t(yes, please use forward slashes)")
	fmt.Println("Example for absolute an path on Linux/Unix:\n\t/home/yourusername/public")
	fmt.Print("Root directory? ")
	rootDirectory := mutils.ReadStdinLine()
	rootDirectory = mutils.BackToForwardSlashes(rootDirectory)
	for strings.HasSuffix(rootDirectory, "/") {
		// trimming the ending slashes from path
		rootDirectory = rootDirectory[0:(len(rootDirectory) - 1)]
	}

	finfo, err := os.Stat(rootDirectory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			mutils.FatalMessage("Error: directory " + rootDirectory + " does not exist.")
		} else {
			mutils.FatalError("error while checking root directory \""+rootDirectory+"\"", err)
		}
	}
	if !finfo.IsDir() {
		mutils.FatalMessage("Error: " + rootDirectory + " is not a directory.")
	}

	// final check
	fmt.Println()
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Final check. Please review the values before saving them")
	fmt.Println("in the configuration directory.")
	fmt.Println(" - configuration directory: \"" + directory + "\"")
	fmt.Println(" - root directory: \"" + rootDirectory + "\"")
	fmt.Println(" - administrator username: \"" + adminUsername + "\"")
	fmt.Print(" - administrator password: ")
	for i := 0; i < len(adminPassword); i++ {
		fmt.Print("*")
	}
	fmt.Println()
	fmt.Println(" - HTTP port: \"" + strconv.Itoa(port) + "\"")
	fmt.Println()
	if !userDefinedConfigDir {
		fmt.Println("IMPORTANT NOTICE")
		fmt.Println("Note that by default the configuration directory is contained in")
		fmt.Println("the directory where the executable is located.")
		fmt.Println("If you want to change this directory, run the server using the")
		fmt.Println("optional with -c argument.")
		fmt.Println("Example for Linux:")
		fmt.Println("  httpiccolo -c \"/home/user/.config/httpiccolo\"")
		fmt.Println("Example for Windows:")
		fmt.Println("  httpiccolo -c \"C:/webtools/httpiccolo/config\"")
		fmt.Println()
	}
	fmt.Print("Do you confirm [y/n]? ")
	confirm := mutils.ReadStdinLine()

	// save configuration
	if strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes") {
		err := os.Mkdir(directory, 0750)
		if err != nil && !os.IsExist(err) {
			mutils.FatalError("fatal error", err)
		}
		if os.IsExist(err) {
			mutils.FatalMessage("Error: the directory \"" + directory + "\" exists,\nif you want to reset the configuration, delete it.")
		}
		users = make(map[string]string, 10)
		configuration = make(map[string]string, 10)
		permissions = make(map[string]string, 10)
		users[adminUsername] = mutils.HashPassword(adminPassword)
		configuration["admin_path"] = "admin"
		configuration["admin_users"] = adminUsername
		configuration["root_directory"] = rootDirectory
		configuration["http_port"] = strconv.Itoa(port)
		mdao.WriteUsersJson(directory, users)
		mdao.WritePermissionsJson(directory, permissions)
		mdao.WriteGeneralParametersJson(directory, configuration)
		fmt.Println("Configuration saved!\nPlease restart...")
		time.Sleep(6 * time.Second)
		os.Exit(0)
	} else {
		mutils.FatalMessage("Configuration not saved.")
		os.Exit(0)
	}
}
