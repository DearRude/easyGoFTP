package ftpserver

import (
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
)

var certCacheDir = "./certs"

func GetTLSConfig(domain string) *tls.Config {
	m := autocert.Manager{
		Cache:      autocert.DirCache(certCacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}

	return &tls.Config{
		GetCertificate: m.GetCertificate,
		MinVersion:     tls.VersionTLS12, // improves cert reputation score at https://www.ssllabs.com/ssltest/
	}
}
