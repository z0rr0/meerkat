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

// Package main implements client part of Meerkat project.
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
)

// Server is main server configuration.
type Server struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	PublicKey string `json:"public_key"`
	publicKey *rsa.PublicKey
	udpConn   *net.UDPConn
}

// Service is client service struct.
type Service struct {
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

// Load initializes client configuration.
func (c *Config) Load(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return errors.New("failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}
	c.Server.publicKey = key
	udpConn, err := net.DialUDP("udp", nil, c.Server.UDPAddr())
	if err != nil {
		return err
	}
	c.Server.udpConn = udpConn
	return nil
}

// send write udp message to remove server.
func (s *Server) send(msg []byte) error {
	n, err := s.udpConn.Write(msg)
	if err != nil {
		return err
	}
	loggerInfo.Printf("wrote %v bytes\n", n)
	return nil
}

// UDPAddr returns server udp address.
func (s *Server) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: net.ParseIP(s.Host), Port: s.Port}
}
