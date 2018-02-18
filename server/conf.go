package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	dbSessionKey key = "db_session"
)

// key is internal type for context types.
type key string

// MongoCfg is database configuration settings
type MongoCfg struct {
	Hosts      []string `json:"hosts"`
	Port       uint     `json:"port"`
	Timeout    uint     `json:"timeout"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Database   string   `json:"database"`
	AuthDB     string   `json:"authdb"`
	ReplicaSet string   `json:"replica"`
	Ssl        bool     `json:"ssl"`
	SslKeyFile string   `json:"sslkeyfile"`
	Reconnects int      `json:"reconnects"`
	RcnTime    int64    `json:"rcntime"`
	PoolLimit  int      `json:"poollimit"`
	Debug      bool     `json:"debug"`
	MongoCred  *mgo.DialInfo
	Logger     *log.Logger
}

// WebAdmin is admin web interface configuration.
type WebAdmin struct {
	Host string `json:"host"`
	Port uint   `json:"port"`
}

// Server is main server configuration.
type Server struct {
	Host       string `json:"host"`
	Port       uint   `json:"port"`
	PrivateKey string `json:"private_key"`
	privateKey *rsa.PrivateKey
}

// Config is main configuration info.
type Config struct {
	WebAdmin WebAdmin `json:"web_admin"`
	Server   Server   `json:"server"`
	Db       MongoCfg `json:"database"`
}

// Addresses returns an array of available MongoDB connections addresses.
func (cfg *MongoCfg) Addresses() []string {
	hosts := make([]string, len(cfg.Hosts))
	port := fmt.Sprint(cfg.Port)
	for i, host := range cfg.Hosts {
		hosts[i] = net.JoinHostPort(host, port)
	}
	return hosts
}

func (cfg *MongoCfg) credential() error {
	if cfg.Ssl {
		pool := x509.NewCertPool()
		pemData, err := ioutil.ReadFile(cfg.SslKeyFile)
		if err != nil {
			return err
		}
		ok := pool.AppendCertsFromPEM(pemData)
		if !ok {
			return errors.New("invalid certificate")
		}
		cert, err := tls.X509KeyPair(pemData, pemData)
		if err != nil {
			return err
		}
		tlsConfig := &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cert},
		}
		dial := func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			if err != nil {
				cfg.Logger.Printf("tls.Dial(%s) failed with %v", addr, err)
				return nil, err
			}
			cfg.Logger.Printf("SSL connection: %v", addr.String())
			return conn, nil
		}
		cfg.MongoCred = &mgo.DialInfo{
			Addrs:          cfg.Addresses(),
			Timeout:        time.Duration(cfg.Timeout) * time.Second,
			Database:       cfg.Database,
			Source:         cfg.AuthDB,
			Username:       cfg.Username,
			Password:       cfg.Password,
			ReplicaSetName: cfg.ReplicaSet,
			DialServer:     dial,
		}
	} else {
		cfg.MongoCred = &mgo.DialInfo{
			Addrs:          cfg.Addresses(),
			Timeout:        time.Duration(cfg.Timeout) * time.Second,
			Database:       cfg.Database,
			Source:         cfg.AuthDB,
			Username:       cfg.Username,
			Password:       cfg.Password,
			ReplicaSetName: cfg.ReplicaSet,
		}
	}
	return nil
}

// Close releases configuration resources.
func (c *Config) Close(ctx context.Context) {
	session, err := CtxGetDBSession(ctx, false)
	if err == nil {
		session.Close()
	}
}

// DbConnect sets database connection.
func (c *Config) DbConnect(ctx context.Context) (context.Context, error) {
	session, err := CtxGetDBSession(ctx, true)
	if err == nil {
		return ctx, nil
	}
	if session != nil {
		session.Close()
	}
	err = c.Db.credential()
	if err != nil {
		return ctx, err
	}
	session, err = mgo.DialWithInfo(c.Db.MongoCred)
	if err != nil {
		return ctx, err
	}
	if c.Db.PoolLimit > 1 {
		session.SetPoolLimit(c.Db.PoolLimit)
	}
	if c.Db.Debug {
		mgo.SetLogger(loggerInfo)
		mgo.SetDebug(true)
	}
	return CtxSetDBSession(ctx, session), nil
}

// CtxSetDBSession saves db session object to the context.
func CtxSetDBSession(ctx context.Context, s *mgo.Session) context.Context {
	return context.WithValue(ctx, dbSessionKey, s)
}

// CtxGetDBSession finds and returns MongoDB session from the Context.
func CtxGetDBSession(ctx context.Context, sendPing bool) (*mgo.Session, error) {
	s, ok := ctx.Value(dbSessionKey).(*mgo.Session)
	if !ok {
		return nil, errors.New("not found context db session")
	}
	if sendPing {
		return s, s.Ping()
	}
	return s, nil
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
	data, err := ioutil.ReadFile(cfg.Server.PrivateKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	pub, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	cfg.Server.privateKey = pub
	return cfg, nil
}

// GenKeys generates and prints new RSA keys pair.
func GenKeys(bits int, logger *log.Logger) {
	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		logger.Fatalln(err)
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}
	if err := pem.Encode(os.Stdout, privateKeyPEM); err != nil {
		logger.Fatalln(err)
	}
	publicKeyPEM := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&pk.PublicKey)}
	if err := pem.Encode(os.Stdout, publicKeyPEM); err != nil {
		logger.Fatalln(err)
	}
}
