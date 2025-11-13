package types

import (
	"crypto/tls"
)

type MysqlConfig struct {
	Name      string
	User      string
	Pass      string
	Host      string
	Port      string
	TLS       bool
	TLSConfig *tls.Config
}
