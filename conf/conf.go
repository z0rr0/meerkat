// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package conf contains interface and methods to init configurations.
package conf

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Loader is interface for configuration load.
type Loader interface {
	Load([]byte) error
}

// ReadFile reads file data.
func ReadFile(l Loader, name string) error {
	absPath, err := filepath.Abs(strings.Trim(name, " "))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil
	}
	return l.Load(data)
}
