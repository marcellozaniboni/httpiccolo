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
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"marcellozaniboni.net/httpiccolo/mdao"
	"marcellozaniboni.net/httpiccolo/mstatic"
	"marcellozaniboni.net/httpiccolo/mutils"
)

// this will be inserted as a comment in admin pages
const httpiccoloVersion string = "0.9.dev"

// this is used for version checking
const httpiccoloProductId string = "5d72b94e-255f-48d9-b2a9-0d3d2ef97df5"

// configuration contains the main configuration
var configuration map[string]string

// users contains a username-password map for configured users
var users map[string]string

// permissions contains a directory-userlist map for restricted access control
var permissions map[string]string

// configpath directory contains the json files where configuration is stored
var configpath string

// restartneeded true means that the system complains untill restarted
var restartneeded bool = false

// userDefinedConfigDir is true when -c command line argument is used
var userDefinedConfigDir = false

// httpGenaralHandler handles all the requests
func httpGenaralHandler(w http.ResponseWriter, r *http.Request) {
	httppath := r.URL.Path
	switch httppath {
	case "/" + configuration["admin_path"]:
		webadminconsole(w, r) // display the administrator console
	case "/" + configuration["admin_path"] + "/save_config":
		websaveconfigurationaction(w, r) // action for saving configuration
	case "/" + configuration["admin_path"] + "/change_password":
		webchangepasswordaction(w, r) // action for changing the password to a single username
	case "/" + configuration["admin_path"] + "/new_user":
		webnewuseraction(w, r) // action for new user creation
	case "/" + configuration["admin_path"] + "/delete_user":
		webdeleteuseraction(w, r) // action for deleting a user
	case "/" + configuration["admin_path"] + "/new_perm":
		webnewpermaction(w, r) // action for new permission
	case "/" + configuration["admin_path"] + "/new_perm_form":
		webnewpermform(w, r) // web page for configuring a new permission
	case "/" + configuration["admin_path"] + "/change_perm":
		webchangepermaction(w, r) // action for changing the userlist to a single permission
	case "/" + configuration["admin_path"] + "/delete_perm":
		webdeleteperm(w, r) // action for deleting a single permission
	case "/login_action":
		webloginaction(w, r) // action called by the login form
	case "/favicon.ico":
		w.Header().Set("Content-Disposition", "attachment; filename=favicon.ico")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(mstatic.ImgFavicon32ico)
	case "/favicon-32.png":
		w.Header().Set("Content-Disposition", "attachment; filename=favicon-32.png")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(mstatic.ImgFavicon32png)
	case "/favicon-128.png":
		w.Header().Set("Content-Disposition", "attachment; filename=favicon-128.png")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(mstatic.ImgFavicon128png)
	case "/favicon-180.png":
		w.Header().Set("Content-Disposition", "attachment; filename=favicon-180.png")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(mstatic.ImgFavicon180png)
	case "/favicon-192.png":
		w.Header().Set("Content-Disposition", "attachment; filename=favicon-192.png")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(mstatic.ImgFavicon192png)
	default:
		// directory/file browsing
		webgenericbrowsing(w, r)
	}
}

func showLicenseTerms() {
	fmt.Println(` This program is free software; you can redistribute it and/or
 modify it under the terms of the GNU General Public License as
 published by the Free Software Foundation, version 2.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public
 License along with this program; if not, you can visit
 https://www.gnu.org/licenses/
 or write to Free Software Foundation, Inc., 675 Mass Ave,
 Cambridge, MA 02139, USA.`)
}

func main() {
	fmt.Print(terminalTitle)

	configpathPtr := flag.String("c", "", "use a custom configuration directory\nif it doesn't exist, it will be created\nif it exists, it must already contain config files")
	licensePtr := flag.Bool("l", false, "show license terms and exit")
	checkNewVersionPtr := flag.Bool("v", false, "check for a new version")
	// printConfigPtr := flag.Bool("p", false, "print current configuration and exit")

	flag.Parse()

	if *licensePtr {
		showLicenseTerms()
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	if *checkNewVersionPtr {

		resp, err := http.Get("https://www.marcellozaniboni.net/release/" + httpiccoloProductId)
		if err != nil {
			mutils.FatalError("Cannot run http request", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("Sorry, I cannot check for new a version.")
			fmt.Println("HTTP response status:", resp.Status)
			time.Sleep(4 * time.Second)
			os.Exit(1)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			mutils.FatalError("Cannot read http response", err)
		}
		lastVersion := string(body)

		if lastVersion == httpiccoloVersion {
			fmt.Println("This is the last available official version.")
		} else {
			fmt.Println("Current release: " + httpiccoloVersion + ", last official release: " + lastVersion)
		}

		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	// check for configuration directory
	if *configpathPtr == "" {
		// default path
		exepath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			mutils.FatalError("fatal error", err)
		}
		exepath = mutils.BackToForwardSlashes(exepath)
		configpath = exepath + "/settings"
	} else {
		// user-defined configuration directory
		userDefinedConfigDir = true
		configpath = *configpathPtr
		configpath = mutils.BackToForwardSlashes(configpath)
		for strings.HasSuffix(configpath, "/") {
			// trimming the ending slashes from path
			configpath = configpath[0:(len(configpath) - 1)]
		}
	}
	finfo, err := os.Stat(configpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configWizard(configpath)
		} else {
			mutils.FatalError("error while checking configuration directory \""+configpath+"\"", err)
		}
	}
	if !finfo.IsDir() {
		mutils.FatalMessage("error: " + configpath + " is not a directory")
	}

	fmt.Println("Configuration directory:\n\t" + configpath)
	if *configpathPtr == "" {
		fmt.Println("\t(You can run the server with -c argument\n\tto define a different directory.)")
		fmt.Println()
	}

	// load configuration
	configuration = mdao.ReadGeneralParameters(configpath)
	users = mdao.ReadUsers(configpath)
	permissions = mdao.ReadPermissions(configpath)

	fmt.Println("The HTTP server has started:\n\thttp://localhost:" + configuration["http_port"] + "/")
	fmt.Println("Administration console:\n\thttp://localhost:" + configuration["http_port"] + "/" + configuration["admin_path"] + "\n")

	// start the server
	http.HandleFunc("/", httpGenaralHandler)
	log.Fatal(http.ListenAndServe(":"+configuration["http_port"], nil))

}

const terminalTitle string = `  _     _   _         _               _
 | |__ | |_| |_ _ __ (_) ___ ___ ___ | | ___
 | '_ \| __| __| '_ \| |/ __/ __/ _ \| |/ _ \
 | | | | |_| |_| |_) | | (_| (_| (_) | | (_) |
 |_| |_|\__|\__| .__/|_|\___\___\___/|_|\___/
               |_|

  a small file server via http - version ` + httpiccoloVersion + `
      copyright © 2022 Marcello Zaniboni

`
