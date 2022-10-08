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
	"log"
	"net/http"
	"strconv"
	"strings"

	"marcellozaniboni.net/httpiccolo/mdao"
	"marcellozaniboni.net/httpiccolo/msession"
	"marcellozaniboni.net/httpiccolo/mstatic"
	"marcellozaniboni.net/httpiccolo/mutils"
)

// verifyLoggedUser returns tre username and also checks if
// an administrator user is
// logged in; warning: this function uses session and cookies,
// so call it before writing the response. It returns the
// username and a flag that means administrator=true/false.
// Being read from the session, username can be "".
func verifyLoggedUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	accessGranted := false // by default nobody can
	// read logged username
	session := msession.GetSession(w, r)
	username := session.Get("username")
	session.Save()
	// verify if it's an administrator
	administrators := strings.Split(configuration["admin_users"], ",")
	for _, administrator := range administrators {
		if administrator == username {
			accessGranted = true
			break
		}
	}
	return username, accessGranted
}

func websaveconfigurationaction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("save configuration action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		log.Println("admin page - save configuration, ", r.Form)
		for k, v := range r.Form {
			configuration[k] = v[0]
		}
		mdao.WriteGeneralParametersJson(configpath, configuration)
		restartneeded = true
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - saving configuration", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webchangepasswordaction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("change password action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - change password, ", form)
		var u, p string
		for k, v := range form {
			if k == "change_password_usr" {
				u = v[0]
			} else if k == "change_password_pwd" {
				p = v[0]
			}
		}
		if u != "" && p != "" {
			// save only valid users
			users[u] = mutils.HashPassword(p)
			mdao.WriteUsersJson(configpath, users)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - changing password", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webdeleteuseraction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("delete user action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - delete user, ", form)
		var u string
		for k, v := range form {
			if k == "delete_user_usr" {
				u = v[0]
			}
		}
		if u != "" {
			// delete user
			delete(users, u)
			mdao.WriteUsersJson(configpath, users)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - deleting user", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webnewuseraction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("new user action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - new user, ", form)
		var u, p string
		for k, v := range form {
			if k == "new_user_usr" {
				u = v[0]
			} else if k == "new_user_pwd" {
				p = v[0]
			}
		}
		if u != "" && p != "" {
			// save only valid users
			// note that if the username already exists, the existing item is overwritten
			users[u] = mutils.HashPassword(p)
			mdao.WriteUsersJson(configpath, users)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - new user", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webnewpermaction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("new permission action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - new perm, ", form)
		var path, ulist string
		for k, v := range form {
			if k == "new_perm_path" {
				path = v[0]
			} else if k == "new_perm_userlist" {
				ulist = v[0]
			}
		}
		// save only valid permissions
		if path != "" && ulist != "" {
			// note: if the path already exist, the user list will be overwritten
			permissions[path] = ulist
			mdao.WritePermissionsJson(configpath, permissions)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - new permission", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webchangepermaction(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("change permission action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - change perm, ", form)
		var path, ulist string
		for k, v := range form {
			if k == "change_perm_path" {
				path = v[0]
			} else if k == "change_perm_userlist" {
				ulist = v[0]
			}
		}
		if path != "" && ulist != "" {
			// save only valid users
			permissions[path] = ulist
			mdao.WritePermissionsJson(configpath, permissions)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - change permission", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webdeleteperm(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("delete permission action, access denied for user \"" + username + "\"")
		fmt.Fprintln(w, "access denied for user \""+username+"\"")
	} else {
		r.ParseForm()
		form := r.Form
		log.Println("admin page - delete perm, ", form)
		var path string
		for k, v := range form {
			if k == "delete_perm_path" {
				path = v[0]
			}
		}
		if path != "" {
			// delete user
			delete(permissions, path)
			mdao.WritePermissionsJson(configpath, permissions)
		}
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - deleting permission", false, restartneeded, false))
		fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction("/"+configuration["admin_path"]))
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webadminconsole(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("admin page, access denied for user \"" + username + "\"")
		webloginform(w, r, username)
	} else {
		log.Println("admin page, user \"" + username + "\"")
		fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings", false, restartneeded, false))
		html := strings.Replace(mstatic.HtmlAdminBody, "[root_directory]", configuration["root_directory"], 1)
		html = strings.Replace(html, "[http_port]", configuration["http_port"], 1)
		html = strings.Replace(html, "[admin_path]", configuration["admin_path"], 1)
		html = strings.Replace(html, "[admin_users]", configuration["admin_users"], 1)
		html = strings.Replace(html, "[valign]", "style='vertical-align: middle'", -1)
		html = strings.Replace(html, "[userlist]", mstatic.GetHtmlUserTable(users), 1)
		html = strings.Replace(html, "[permissionlist]", mstatic.GetHtmlPermissionTable(permissions), 1)
		html = strings.Replace(html, "[save_config_action]", "/"+configuration["admin_path"]+"/save_config", 1)
		html += mstatic.HtmlAdminJavascriptAndHiddenForms
		html = strings.Replace(html, "[change_password_action]", "/"+configuration["admin_path"]+"/change_password", 1)
		html = strings.Replace(html, "[new_user_action]", "/"+configuration["admin_path"]+"/new_user", 1)
		html = strings.Replace(html, "[delete_user_action]", "/"+configuration["admin_path"]+"/delete_user", 1)
		html = strings.Replace(html, "[new_perm_form_url]", "/"+configuration["admin_path"]+"/new_perm_form"+"?nonache="+mutils.RandomId(noCacheIdLength), 1)
		html = strings.Replace(html, "[change_permusers_action]", "/"+configuration["admin_path"]+"/change_perm", 1)
		html = strings.Replace(html, "[delete_perm_action]", "/"+configuration["admin_path"]+"/delete_perm", 1)
		fmt.Fprintln(w, html)
		fmt.Fprintln(w, "<!-- httpiccolo version "+httpiccoloVersion+" -->")
		fmt.Fprintln(w, mstatic.HtmlFooter)
	}
}

func webnewpermform(w http.ResponseWriter, r *http.Request) {
	username, isAdmin := verifyLoggedUser(w, r)
	if !isAdmin {
		// login needed
		log.Println("new permission form, access denied for user \"" + username + "\"")
		webloginform(w, r, username)
		return
	}
	log.Println("new permission form, user \"" + username + "\"")
	fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - settings - new private directory", false, restartneeded, false))
	var rootpath string = configuration["root_directory"]
	for strings.HasSuffix(rootpath, "/") || strings.HasSuffix(rootpath, "\\") {
		// trimming the ending slashes from path
		rootpath = rootpath[0:(len(rootpath) - 1)]
	}
	directories, err := mutils.DirTree(rootpath)
	directoryCount := len(directories)
	if directoryCount <= 4096 {
		log.Println("number of available directories:", directoryCount)
	} else {
		log.Println("WARNING! The number of directories is very high:", directoryCount)
	}

	if err != nil {
		fmt.Fprintln(w, "ERROR: one or more directories under the root directory are not readable.")
		fmt.Fprintln(w, "Reconfigure root directory and try again.<br/>")
		fmt.Fprintln(w, err)
		fmt.Fprintln(w, "<br/>&nbsp;<br/><a href=\"/"+configuration["admin_path"]+"?nonache="+mutils.RandomId(noCacheIdLength)+"\">Go back to the settings</a>")
	} else {
		fmt.Fprintln(w, "<form id=\"blablabla\" name=\"blablabla\" action=\"/admin/new_perm_blablabla\" method=\"post\">")
		fmt.Fprintln(w, "<p>Select the users that will access the private directory:</p>")
		fmt.Fprintln(w, "<table class=\"w3-table-all\">")
		fmt.Fprintln(w, "<tr><th>configured user</th></tr>")
		var id int = 0
		for u := range users {
			fmt.Fprintln(w, "<tr><td><input type=\"checkbox\" id=\"usr_"+strconv.Itoa(id)+"\" name=\"usr_"+strconv.Itoa(id)+"\" value=\""+u+"\"/>")
			fmt.Fprintln(w, "<label for=\"usr_"+strconv.Itoa(id)+"\">"+u+"</label></td></tr>")
			id++
		}
		fmt.Fprintln(w, "</table>")
		fmt.Fprintln(w, "<p>Select the directory:</p>")
		fmt.Fprintln(w, "<table class=\"w3-table-all\">")
		// TODO: dir true tree display. Since directories and subdirectories
		// are ordered alphabetically, comparing the number of slashes between
		// an element and the previous on will tell what is their relation:
		//   count(current) == count(previous) ⇒ brother
		//   count(current) > count(previous) ⇒ current is son
		//   count(current) < count(previous) ⇒ current is aunt
		fmt.Fprintln(w, "<tr><th>available directories</th></tr>")
		id = 0
		for _, d := range directories {
			pathLength := strings.LastIndex(d, "/")
			fmt.Fprintln(w, "<tr>")
			fmt.Fprintln(w, "<td><input type=\"radio\" id=\"dir_"+strconv.Itoa(id)+"\" name=\"directory\" value=\"/"+d+"\"/>")
			fmt.Fprintln(w, "&nbsp;<label for=\"dir_"+strconv.Itoa(id)+"\">")
			if pathLength > 0 {
				fmt.Fprint(w, "<font color=\"#707070\">"+d[:pathLength+1]+"</font>")
				fmt.Fprintln(w, d[pathLength+1:]+"</label></td>")
			} else {
				fmt.Fprintln(w, d+"</label></td>")
			}
			fmt.Fprintln(w, "</tr>")
			id++
		}
		fmt.Fprintln(w, "</table>")
		fmt.Fprintln(w, "</form>")
	}
	fmt.Fprintln(w, "<div style=\"margin-top: 16px; margin-bottom: 6px\" align=\"center\"><a href=\"../"+configuration["admin_path"]+"?nonache="+mutils.RandomId(noCacheIdLength)+"\">Cancel (bo back)</a>")
	fmt.Fprintln(w, "&nbsp;&nbsp;&nbsp;<a href=\"#\" onclick=\"createPermission()\">Save</a></div>")
	fmt.Fprintln(w, `
<form id="new_perm_form" name="new_perm_form" action="/`+configuration["admin_path"]+`/new_perm" method="post">
<input id="new_perm_path" name="new_perm_path" type="hidden" value=""/>
<input id="new_perm_userlist" name="new_perm_userlist" type="hidden" value=""/></form>
<script type="text/javascript" charset="utf-8">
function createPermission() {
	var userlist = "";
	for (var i = 0; i < `+strconv.Itoa(len(users))+`; i++) {
		if (document.getElementById("usr_" + i).checked) {
			if (userlist != "") userlist += ",";
			userlist += document.getElementById("usr_" + i).value;
		}
	}
	var path = "";
	for (var i=0; i < `+strconv.Itoa(directoryCount)+`; i++) {
		if (document.getElementById("dir_" + i).checked) {
			path = document.getElementById("dir_" + i).value;
			break;
		}
	}
	if (userlist == "") {
		alert("Select one or more users.");
		return;
	}
	if (path == "") {
		alert("Select a directory.");
		return;
	}
	document.getElementById("new_perm_path").value = path;
	document.getElementById("new_perm_userlist").value = userlist;
	document.getElementById("new_perm_form").submit();
}
</script>`)

	fmt.Fprintln(w, "<!-- httpiccolo version "+httpiccoloVersion+" -->")
	fmt.Fprintln(w, mstatic.HtmlFooter)
}
