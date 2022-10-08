package mdao

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
	"encoding/json"
	"io"
	"log"
	"os"
)

/////////////////
// PERMISSIONS //
/////////////////

// JsonPermission is an json item of a configured permission
type JsonPermission struct {
	Directory string `json:"directory"`
	Userlist  string `json:"userlist"`
}

// JsonPermList is a json collection of JsonPermission items
type JsonPermList struct {
	Permissions []JsonPermission `json:"permissions"`
}

func ReadPermissions(configpath string) map[string]string {
	// load the configuration
	var cfg JsonPermList
	filename := configpath + "/permissions.json"
	configfile, err := os.Open(filename)
	if err != nil {
		log.Println("The configuration directory exists, but it does not contain \"" + filename + "\"; if you want to reset the configuration, remove the entire directory, not just its files.")
		log.Fatal(err)
	}
	defer configfile.Close()
	filecontent, err := io.ReadAll(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(filecontent, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// build the map that will be returned
	var configMap = map[string]string{}
	for _, v := range cfg.Permissions {
		configMap[v.Directory] = v.Userlist
	}
	return configMap
}

func WritePermissionsJson(path string, permissions map[string]string) {
	var jperms JsonPermList
	var jpermSlice []JsonPermission
	for k, v := range permissions {
		var jp JsonPermission
		jp.Directory = k
		jp.Userlist = v
		jpermSlice = append(jpermSlice, jp)
	}
	jperms.Permissions = jpermSlice

	json, err := json.MarshalIndent(jperms, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	filename := path + "/permissions.json"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n, err := f.Write(json)
	if err != nil {
		log.Fatal(err)
	}
	if n == 0 {
		log.Fatal("error: could not write anything to", filename)
	}
}
