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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"marcellozaniboni.net/httpiccolo/mstatic"
	"marcellozaniboni.net/httpiccolo/mutils"
)

const noCacheIdLength int = 10

// webgenericbrowsing provides the html pages for browsing and serving
// directories and files to the web browsers.
func webgenericbrowsing(w http.ResponseWriter, r *http.Request) {
	// user identification
	username, isAdmin := verifyLoggedUser(w, r) // admin flag ignored for now

	// manage login if requested by the user
	r.ParseForm()
	if r.Form.Get("login") == "spontaneous" {
		log.Println("login required by user")
		webloginform(w, r, username)
		return
	}

	// httppath is the logical web path
	var httppath string = r.URL.Path
	// rootpath on the filesystem for published contents
	var rootpath string = configuration["root_directory"]
	for strings.HasSuffix(httppath, "/") || strings.HasSuffix(httppath, "\\") {
		// trimming the ending slashes from path
		httppath = httppath[0:(len(httppath) - 1)]
	}
	for strings.HasSuffix(rootpath, "/") || strings.HasSuffix(rootpath, "\\") {
		// trimming the ending slashes from path
		rootpath = rootpath[0:(len(rootpath) - 1)]
	}

	// resourcepath is the physical filesystem path
	var resourcepath string = rootpath + "/" + httppath
	resourcepath = strings.ReplaceAll(resourcepath, "//", "/")
	log.Print("file/dir browsing: \""+httppath+"\" >>> \"", resourcepath+"\" - user \""+username+"\"")

	// filesystem search
	info, err := os.Stat(resourcepath)
	if err != nil {
		time.Sleep(4 * time.Second) // penalty time
		fmt.Fprintln(w, mstatic.ErrNoContent)
		log.Print("nothing found for \"" + httppath + "\"")
		return
	}

	// check if the requested resource is private and in this case check if
	// the logged user is permitted to get it and if not, redirect to the
	// login form, setting the requested resource as the redirect URL
	isPrivate := false
	noGrantFound := true
	for privateDirectory, allowedUsers := range permissions {
		if strings.HasPrefix(httppath, privateDirectory) {
			isPrivate = true
			allowedUserSlice := strings.Split(allowedUsers, ",")
			for _, u := range allowedUserSlice {
				if username == u {
					noGrantFound = false
					break
				}
			}
		}
	}
	if isPrivate && noGrantFound {
		// the directory is private and the user is not allowed >>> login form
		log.Println("access denied for user " + username + " to " + httppath)
		webloginform(w, r, username)
		return
	}

	// files must be served directly while directories must be browser
	if info.IsDir() {
		// info contains the elements inside the directory
		infos, err := os.ReadDir(resourcepath)
		if err != nil {
			fmt.Fprintln(w, "directory browsing error, contact the system administrator")
			log.Println("directory browsing error:", err)
		}

		// the web page header
		title := "Contents of " // web page big title
		if httppath == "" {
			title += "/"
		} else {
			title += httppath
		}
		htmlHeader := mstatic.GetHtmlHeader(title, true, restartneeded, true)
		if username == "" {
			htmlHeader = strings.ReplaceAll(htmlHeader, "[logged_username]", "<span class=\"w3-text-dark-grey\"><i>anonymous</i></span>")
		} else {
			if isAdmin {
				htmlHeader = strings.ReplaceAll(htmlHeader, "[logged_username]", "<span class=\"w3-text-red\">"+username+"</span>")
			} else {
				htmlHeader = strings.ReplaceAll(htmlHeader, "[logged_username]", username)
			}
		}
		fmt.Fprintln(w, htmlHeader)

		// table containing the fines inside the directory
		fmt.Fprintln(w, "<table class='w3-table-all'>\n<tr><th>name</th><th>size</th><th>time</th></tr>")
		if httppath != "" { // link to the partent directory ".."
			var parentDirectoryHttpPath string
			i := strings.LastIndex(httppath, "/")
			if i > 0 {
				parentDirectoryHttpPath = httppath[:i]
			} else {
				parentDirectoryHttpPath = "/"
			}
			fmt.Fprintln(w, "<tr class='w3-hover-text-brown'>")
			fmt.Fprint(w, "<td title='open parent directory'><font color='#666666'>&uuarr;</font>")
			fmt.Fprint(w, "<a href='"+parentDirectoryHttpPath+"?nonache="+mutils.RandomId(noCacheIdLength)+"'><b>&nbsp;..&nbsp;</b></a>")
			fmt.Fprint(w, "<font color='#666666'>&uuarr;</font></a></td>")
			fmt.Fprintln(w, "<td>-</td><td>-</td></tr>")
		}

		// stat counters will be printed under the table
		fileCounter := 0
		dirCounter := 0
		var fileSizeSum int64 = 0
		for _, f := range infos { // main loop for sub-directories
			if f.IsDir() {
				// show restricted access info
				var checkKey string
				if httppath == "" {
					checkKey = "/" + f.Name()
				} else {
					checkKey = httppath + "/" + f.Name()
				}
				_, lockedDir := permissions[checkKey]

				if lockedDir && username == "" {
					// lot logged users cannot see private directory names
					log.Println("private directory name " + f.Name() + " hidden for anonymous users")
				} else {
					fmt.Fprintln(w, "<tr class='w3-hover-text-brown'>")
					fmt.Fprintln(w, "<td>[<a href='"+httppath+"/"+f.Name()+"?nonache="+mutils.RandomId(noCacheIdLength)+"'>"+f.Name()+"]</a></td>")
					fmt.Fprint(w, "<td><small><i>directory")
					if lockedDir {
						fmt.Fprint(w, " [PRIVATE]")
					}
					fmt.Fprintln(w, "</small></i></td>")
					dirCounter++
					modificationTime := "???"
					fileinfo, err := f.Info()
					if err != nil {
						log.Println("error reading dir info for " + f.Name())
					} else {
						modificationTime = fileinfo.ModTime().Format("2006-01-02 15:04:05")
					}
					fmt.Fprintln(w, "<td>"+modificationTime+"</td>")
					fmt.Fprintln(w, "</tr>")
				}
			}
		}
		for _, f := range infos { // main loop files
			if !f.IsDir() {
				var fileSize int64 = 0
				modificationTime := "???"
				fileinfo, err := f.Info()
				if err != nil {
					log.Println("error reading file info for " + f.Name())
				} else {
					modificationTime = fileinfo.ModTime().Format("2006-01-02 15:04:05")
					fileSize = fileinfo.Size()
				}
				fmt.Fprintln(w, "<tr class='w3-hover-text-indigo'>")
				fmt.Fprintln(w, "<td><a href='"+httppath+"/"+f.Name()+"?nocache="+mutils.RandomId(noCacheIdLength)+"'>"+f.Name()+"</a></td>")
				fmt.Fprintln(w, "<td>"+mutils.FormatFileSize(fileSize)+"</td>")
				fileCounter++
				fileSizeSum += fileSize
				fmt.Fprintln(w, "<td>"+modificationTime+"</td>")
				fmt.Fprintln(w, "</tr>")
			}
		}
		fmt.Fprintln(w, "</table>")

		// stats section
		fmt.Fprint(w, "<p>")
		fmt.Fprint(w, dirCounter)
		if dirCounter == 1 {
			fmt.Fprint(w, " directory, ")
		} else {
			fmt.Fprint(w, " directories, ")
		}
		fmt.Fprint(w, fileCounter)
		if fileCounter == 1 {
			fmt.Fprint(w, " file, total size: ")
		} else {
			fmt.Fprint(w, " files, total size: ")
		}
		fmt.Fprintln(w, mutils.FormatFileSize(fileSizeSum)+"</p>")
		fmt.Fprintln(w, mstatic.HtmlFooter)
	} else { // file links are served directly
		s := strings.Split(httppath, "/")
		if len(s) > 0 {
			downloadFileName := s[len(s)-1] // this eliminate file path
			log.Print("downloading: \"" + downloadFileName + "\"")
			// files with text extension are written in the response, other files are downloaded
			if strings.HasSuffix(strings.ToLower(downloadFileName), ".html") || strings.HasSuffix(strings.ToLower(downloadFileName), ".htm") || strings.HasSuffix(strings.ToLower(downloadFileName), ".txt") || strings.HasSuffix(strings.ToLower(downloadFileName), ".md") || strings.HasSuffix(strings.ToLower(downloadFileName), ".log") {
				file, err := os.Open(resourcepath)
				if err != nil {
					fmt.Fprintln(w, "error while opening file, please report to the administrator")
					log.Print("error while opening file "+resourcepath, err)
					return
				}
				defer file.Close()
				buffer := make([]byte, 4096)
				for {
					count, err := file.Read(buffer)
					if err != nil {
						if err != io.EOF {
							fmt.Fprintln(w, "error while reading file, please report to the administrator")
							log.Print("error while reading file "+resourcepath, err)
							return
						}
						break
					}
					if count > 0 {
						w.Write(buffer[:count])
					}
				}
			} else { // other files must be downloaded
				w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(downloadFileName))
				w.Header().Set("Content-Type", "application/octet-stream")
				http.ServeFile(w, r, resourcepath)
			}
		} else {
			// this should never happen
			log.Println("error: unable to determine file name for " + httppath)
		}
	}
}
