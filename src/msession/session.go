package msession

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
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const defaultSessionExpireTime time.Duration = 360 * time.Minute // 6 hours

const defaultMapSize int = 32
const sessionCookieName string = "msessionid"

var sessionRandomNeedsSeed = true

// Session struct contains one real instance of a web Session.
type Session struct {
	id             string
	expiry         time.Time
	items          map[string]string // TODO: turn this into a sync.Map to avoid concurrency for the same client
	responseWriter http.ResponseWriter
}

// sessions is a private map containing session structs.
var sessions sync.Map

// GetSession returns a valid Session object. The servlet client (http
// handler function) can then call Set() and Get() methods to write and
// read key-value pairs to/from the session.
// Important notice: before your client function ends, remember to call
// the method Save() to store the session in the web server memory for
// the next http requests and to refresh the expiration time!
func GetSession(w http.ResponseWriter, r *http.Request) Session {
	sessionGC()   // cleanup expired sessions
	var s Session // this will contain a valid Session after the following "if"
	readcookie, err := r.Cookie(sessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		// the browser has not the cookie
		buildNewSession(&s)
	} else {
		// the browser has the cookie
		stored, found := sessions.Load(readcookie.Value)
		if !found {
			// the browser has the cookie but it's not in sessions
			buildNewSession(&s)
		} else {
			// the browser has the cookie and it's OK
			if typedValue, ok := stored.(Session); ok {
				s = typedValue
			} else {
				log.Fatal("Unexpected sync.Map value type!")
			}
			s.expiry = time.Now().Add(defaultSessionExpireTime)
		}
	}
	s.responseWriter = w
	return s
}

// Set writes a key-value pair in session.
func (s *Session) Set(key string, value string) {
	if s.id != "" {
		s.expiry = time.Now().Add(defaultSessionExpireTime)
		s.items[key] = value // store the value
	}
}

// Get reads a value, given a key. If the value does not exist
// an empty string is returned.
func (s *Session) Get(key string) string {
	if s.id == "" {
		log.Println("invalid session, use GetSession to get a valid instance")
		return ""
	}
	s.expiry = time.Now().Add(defaultSessionExpireTime)
	return s.items[key]
}

// Save updates the session storage in memory and send/updates
// the session cookie in the web browser.
func (s *Session) Save() {
	if s.id != "" {
		sessions.Store(s.id, *s) // TODO mind about the pointer: is it really useful?
		cookie := http.Cookie{Name: sessionCookieName, Value: s.id, Expires: s.expiry}
		cookie.Path = "/"
		http.SetCookie(s.responseWriter, &cookie)
	} else {
		log.Println("invalid session, use GetSession to get a valid instance")
	}
}

// Expirate invalidates the session setting the expiry to time one hour ago.
// After calling Expirate(), remember to call Save(), and not to use Get()
// or Set(), because the session will be invalid.
func (s *Session) Expirate() {
	s.expiry = time.Now().Add(-time.Hour)
}

// buildNewSession creates a new Session and add it to sessions.
func buildNewSession(s *Session) {
	// create session attributes
	s.id = randomSessionId()
	s.expiry = time.Now().Add(defaultSessionExpireTime)
	s.items = make(map[string]string, defaultMapSize)
	// add the session to the container
	sessions.Store(s.id, *s) // TODO mind about the pointer: is it really useful?
}

// sessionGC garbage collects every expired session from sessions.
func sessionGC() {
	sessions.Range(func(parK, parV any) bool {
		var id string
		var s Session
		if skey, ok := parK.(string); ok {
			id = skey
		} else {
			log.Fatal("Unexpected sync.Map key type!")
		}
		if sval, ok := parV.(Session); ok {
			s = sval
		} else {
			log.Fatal("Unexpected sync.Map value type!")
		}
		if s.expiry.Before(time.Now()) {
			sessions.Delete(id)
		}
		return true
	})
}

// PrintSessions does pretty logging (for development purposes).
// It returns a formatted, printable string.
func PrintSessions() string {
	var info string
	var activeSessionCount int = 0
	sessions.Range(func(parK, parV any) bool {
		var id string
		var s Session
		if skey, ok := parK.(string); ok {
			id = skey
		} else {
			log.Fatal("Unexpected sync.Map key type!")
		}
		if sval, ok := parV.(Session); ok {
			s = sval
		} else {
			log.Fatal("Unexpected sync.Map value type!")
		}
		info += fmt.Sprintln("\tcookie:", id)
		info += fmt.Sprintln("\texpiry:", s.expiry.Format("2006-01-02 15:04:05"))
		for k, v := range s.items {
			info += fmt.Sprintln("\t\tkey:", k, "\n\t\tval:", v)
		}
		activeSessionCount++
		return true
	})

	return "number of sessions: " + strconv.Itoa(activeSessionCount) + "\n" + info
}

// randomSessionId returns a random hexadeciman string useful
// as a key for sessions and their cookies-
func randomSessionId() string {
	var randomHexString string
	if sessionRandomNeedsSeed {
		rand.Seed(time.Now().UnixNano())
		sessionRandomNeedsSeed = false
	}

	for i := 0; i < 32; i++ {
		n := rand.Intn(256)
		onebyte := fmt.Sprintf("%X", n)
		if len(onebyte) == 1 {
			onebyte = "0" + onebyte
		}
		randomHexString += onebyte
	}
	return randomHexString
}
