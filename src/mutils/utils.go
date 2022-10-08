package mutils

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
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const fatalErrorSleep int = 4

var randomNeedsSeed = true

// HashPassword returns a string hash
func HashPassword(password string) string {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(password))
	if err != nil {
		FatalError("fatal error", err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// FormatFileSize returns pretty-printed file size
func FormatFileSize(filesize int64) string {
	var retval string
	floatFilesize := float64(filesize)
	switch {
	case floatFilesize < 10000:
		retval = strconv.FormatInt(filesize, 10) + " bytes"
	case floatFilesize < 1024*1024:
		retval = fmt.Sprintf("%.2f Kib", floatFilesize/1024)
	case floatFilesize < 1024*1024*1024:
		retval = fmt.Sprintf("%.2f Mib", floatFilesize/(1024*1024))
	case floatFilesize < 1024*1024*1024*1024:
		retval = fmt.Sprintf("%.2f GiB", floatFilesize/(1024*1024*1024))
	default:
		retval = fmt.Sprintf("%.1f TiB", floatFilesize/(1024*1024*1024*1024))
	}
	return retval
}

// RandomId returns a random string
func RandomId(length int) string {
	if randomNeedsSeed {
		rand.Seed(time.Now().UnixNano())
		randomNeedsSeed = false
	}
	var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randomRunes := make([]rune, length)
	for i := 0; i < length; i++ {
		randomRunes[i] = characters[rand.Intn(len(characters))]
	}
	return string(randomRunes)
}

// FatalMessage prints a fatal message and exits
func FatalMessage(message string) {
	fmt.Println(message)
	time.Sleep(time.Duration(fatalErrorSleep) * time.Second)
	os.Exit(1)
}

// FatalError prints a fatal message and an error and then exits
func FatalError(message string, err error) {
	fmt.Println(message, err)
	time.Sleep(time.Duration(fatalErrorSleep) * time.Second)
	os.Exit(1)
}

// DirTree returns an ordered string slice containing
// recursively the directory names under a root path.
func DirTree(root string) ([]string, error) {
	var directories []string
	var trimRootPath int = len(root)
	e := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && path != root {
				dirname := path[trimRootPath:]
				if dirname[0] == '/' || dirname[0] == '\\' {
					dirname = dirname[1:]
				}
				directories = append(directories, strings.Replace(dirname, "\\", "/", -1))
			}
			return nil
		})
	if e != nil {
		return nil, e
	}
	return directories, nil
}

// ReadStdinLine returns a line read from the stdin (user's keyboard
// in most cases).
func ReadStdinLine() string {
	stdin := bufio.NewReader(os.Stdin)
	input, err := stdin.ReadString('\n')
	if err != nil {
		FatalError("error while readin from standard input: ", err)
	}
	if len(input) > 0 && input[len(input)-1] == '\n' {
		input = input[:len(input)-1]
	}
	if len(input) > 0 && input[len(input)-1] == '\r' {
		input = input[:len(input)-1]
	}
	return input
}

func BackToForwardSlashes(s string) string {
	return strings.ReplaceAll(s, "\\", "/")
}
