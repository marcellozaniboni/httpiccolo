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
	"errors"
	"io"
	"log"
	"os"
)

////////////////////////
// GENERAL PARAMATERS //
////////////////////////

// JsonParam is an json item of general configuration
type JsonParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// JsonConfigList is a json collection of Param items
type JsonConfigList struct {
	Params []JsonParam `json:"params"`
}

func WriteGeneralParametersJson(path string, cfgmap map[string]string) {
	var jcfg JsonConfigList
	var jpar []JsonParam
	for k, v := range cfgmap {
		var jp JsonParam
		jp.Name = k
		jp.Value = v
		jpar = append(jpar, jp)
	}
	jcfg.Params = jpar

	json, err := json.MarshalIndent(jcfg, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	filename := path + "/params.json"
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

func ReadGeneralParameters(configpath string) map[string]string {
	// load the configuration
	var cfg JsonConfigList

	filename := configpath + "/params.json"
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
	for _, v := range cfg.Params {
		configMap[v.Name] = v.Value
	}

	// check value: root directory
	rootDirectory, ok := configMap["root_directory"]
	if !ok {
		log.Fatal("root_directory not defined in " + filename)
	}
	if _, err := os.Stat(rootDirectory); err == nil {
		// fmt.Println("Root directory found:\n\t" + rootDirectory)
	} else if errors.Is(err, os.ErrNotExist) {
		// TODO - this is too violent; fix it in future
		log.Println("root directory not found")
		log.Fatal(err)
	} else {
		log.Println("error checking root directory")
		log.Fatal(err)
	}

	// check value: http port
	_, ok = configMap["http_port"]
	if !ok {
		log.Fatal("http_port not defined in " + filename)
	}

	// check value: admin_path (web management console)
	_, ok = configMap["admin_path"]
	if !ok {
		log.Fatal("admin_path not defined in " + filename)
	}

	// check value: list of admin users
	_, ok = configMap["admin_users"]
	if !ok {
		log.Fatal("admin_users not defined in " + filename)
	}

	return configMap
}
