package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
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

// Config is main configuration info.
type Config struct {
	Admin    WebAdmin `json:"admin"`
	Host     string   `json:"host"`
	Port     uint     `json:"port"`
	Key      string   `json:"key"`
	Db       MongoCfg `json:"database"`
	dbM      sync.Mutex
	released bool
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
	c.dbM.Lock()
	defer c.dbM.Unlock()
	c.released = true

	session, err := CtxGetDBSession(ctx, false)
	if err == nil {
		session.Close()
	}
}

// DbConnect sets database connection.
func (c *Config) DbConnect(ctx context.Context) (context.Context, error) {
	if c.released {
	}
	session, err := CtxGetDBSession(ctx, true)
	if err == nil {
		return ctx, nil
	}
	c.dbM.Lock()
	defer c.dbM.Unlock()

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
