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
	"net"
	"net/http"
	"strings"

	"marcellozaniboni.net/httpiccolo/bruteforce"
	"marcellozaniboni.net/httpiccolo/msession"
	"marcellozaniboni.net/httpiccolo/mstatic"
	"marcellozaniboni.net/httpiccolo/mutils"
)

// webloginform display the web page containing the login form. It is not
// called directly by a user's action. It is called by other http handler
// function when needed (in a sort of server-side redirection).
func webloginform(w http.ResponseWriter, r *http.Request, username string) {
	log.Println("login form")

	// anti brute-force protection
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip != "" {
		if bruteforce.Banned(ip) {
			log.Println("login action: login refused for banned IP " + ip)
			fmt.Fprintln(w, mstatic.ErrBannedIP)
			return
		}
	}

	redirectUrl := r.URL.Path
	for strings.HasSuffix(redirectUrl, "/") {
		redirectUrl = redirectUrl[0:(len(redirectUrl) - 1)]
	}
	if redirectUrl == "" {
		redirectUrl = "/"
	}

	currentUsername := username
	if currentUsername == "" {
		currentUsername = "<i>anonymous</i>"
	}
	fmt.Fprintln(w, mstatic.GetHtmlHeader("", true, restartneeded, false))
	html := strings.Replace(mstatic.HtmlLoginForm, "[current_login_username]", currentUsername, 1)
	html = strings.Replace(html, "[redirect_url]", redirectUrl+"?nonache="+mutils.RandomId(noCacheIdLength), 1)
	fmt.Fprintln(w, html)
	fmt.Fprintln(w, mstatic.HtmlFooter)
}

// webloginaction verifies username and password and if the
// login is successful, sets the username into the session.
// The redirect URL will do the necessary security checks.
func webloginaction(w http.ResponseWriter, r *http.Request) {
	log.Println("login action")

	// anti brute-force protection
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip != "" {
		if bruteforce.Banned(ip) {
			log.Println("login action: login refused for banned IP " + ip)
			fmt.Fprintln(w, mstatic.ErrBannedIP)
			return
		}
	}

	// read login parameters
	r.ParseForm()
	form := r.Form
	var user, pass, url string
	for k, v := range form {
		switch k {
		case "username":
			user = v[0]
		case "password":
			pass = v[0]
		case "redirect_url":
			url = v[0]
		}
	}
	hashedpass := mutils.HashPassword(pass)

	// set the session if the login is ok
	if users[user] == hashedpass {
		s := msession.GetSession(w, r)
		s.Set("username", user)
		s.Save()
	} else {
		// record the failed attempt for brute-force control
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip != "" {
			bruteforce.RecordFailedLogin(ip)
		}
		log.Println("login failed for user \""+user+"\", IP \""+ip+"\", banned =", bruteforce.Banned(ip))
	}
	fmt.Fprintln(w, mstatic.GetHtmlHeader("httpiccolo - loggin in", false, restartneeded, false))
	fmt.Fprintln(w, mstatic.GetHtmlCounterAfterDaoAction(url))
	fmt.Fprintln(w, mstatic.HtmlFooter)

}
