package bruteforce

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
	"strconv"
	"time"

	"marcellozaniboni.net/httpiccolo/mutils"
)

// maxLoginFails is the maximum number of failed logins
// allowed for each IP address; if an IP exceeds this
// limit, in maxLoginsMinutes it is banned. It means that
// in these cases Banned() returns true.
const maxLoginFails int = 5

const maxLoginsMinutes int = 20

type failedAccess struct {
	ip         string
	accessTime time.Time
}

var accessLog = map[string]failedAccess{} // TODO: this should sync.Map!!!

// RecordFailedLogin records a failed login in memory; call it
// when a login fails passing the IP.
func RecordFailedLogin(ip string) {
	const randomKeyLength int = 10

	var fa failedAccess
	fa.ip = ip
	fa.accessTime = time.Now()
	randomKey := mutils.RandomId(randomKeyLength)
	accessLog[randomKey] = fa
}

// cleanOldAccessLogs cleans the failed login logs older
// than maxLoginsMinutes; only useful logs are kept.
func cleanOldAccessLogs() {
	for k, v := range accessLog {
		if v.accessTime.Before(time.Now().Add(-time.Duration(maxLoginsMinutes) * time.Minute)) {
			delete(accessLog, k)
		}
	}
}

// Banned returns the ban status for an IP address: true
// means that the IP is banned.
func Banned(ip string) bool {
	cleanOldAccessLogs()
	var failCount int64 = 0
	for _, v := range accessLog {
		if v.ip == ip {
			failCount++
		}
	}
	return (failCount >= int64(maxLoginFails))
}

// PrintAccessLog does pretty logging (for development purposes).
// It returns a formatted, printable string.
func PrintAccessLog() string {
	info := fmt.Sprintln("number of failed login in the last "+strconv.Itoa(maxLoginsMinutes)+" minutes:", len(accessLog))
	for key, al := range accessLog {
		info += fmt.Sprintln("\tkey =", key)
		info += fmt.Sprintln("\t\ttimestamp =", al.accessTime.Format("2006-01-02 15:04:05"))
		info += fmt.Sprintln("\t\tip =", al.ip)
	}
	return info
}
