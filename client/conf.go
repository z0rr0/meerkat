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
//
// Package main implements client part of Meerkat project.
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Server is main server configuration.
type Server struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	PublicKey string `json:"public_key"`
	publicKey *rsa.PublicKey
}

// Service is client service struct.
type Service struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Exec         string   `json:"exec"`
	Args         []string `json:"args"`
	IgnoreErrors bool     `json:"ignore_errors"`
	Period       int      `json:"period"`
}

// Config is main client configuration info.
type Config struct {
	Server   Server    `json:"server"`
	Services []Service `json:"services"`
}

// readConfigurationFile reads file configuration.
func readConfigurationFile(name string) (*Config, error) {
	cfg := &Config{}
	absPath, err := filepath.Abs(strings.Trim(name, " "))
	if err != nil {
		return cfg, err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return cfg, err
	}
	jsonData, err := ioutil.ReadFile(absPath)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(jsonData, cfg)
	return cfg, err
}

// Configuration reads configuration file and does its validation.
func Configuration(fileName string) (*Config, error) {
	cfg, err := readConfigurationFile(fileName)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(cfg.Server.PublicKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	cfg.Server.publicKey = key
	return cfg, nil
}
