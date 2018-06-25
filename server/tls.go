package server

import (
	"crypto/tls"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/RobertGrantEllis/t9/bindata"
	"github.com/RobertGrantEllis/t9/server/autocert"
)

func (s *server) getTlsConfig() (*tls.Config, error) {

	embeddedCertificate, err := getEmbeddedCertificate()
	if err != nil {
		return nil, err
	}

	// for now, assume no certificates are configured
	getCertificate := func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

		// since this version of the function assumes no other certificates are
		// designated, simply return the embedded certificate
		return embeddedCertificate, nil
	}

	switch {
	case len(s.configuration.CertificateFile) > 0:

		// keyfile is also designated since the configuration normalization checks
		designatedCertificate, err := getDesignatedCertificate(
			s.configuration.CertificateFile,
			s.configuration.KeyFile,
		)
		if err != nil {
			return nil, err
		}

		// a certificate was designated, so override the getCertificate function
		getCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

			if hello.ServerName == `t9` {
				// this is our internal client SNI designation so use embedded cert
				return embeddedCertificate, nil
			}

			// otherwise use the designated cert
			return designatedCertificate, nil
		}

	case s.configuration.UseAutocert:

		autocertManager, err := autocert.New(s.configuration.Autocert, s.logger)
		if err != nil {
			return nil, errors.Wrap(err, `could not instantiate autocert manager`)
		}

		// we are using autocert, so override the getCertificate function
		getCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

			if hello.ServerName == `t9` {
				// this is our internal client SNI designation so use embedded cert
				return embeddedCertificate, nil
			}

			cert, err := autocertManager.GetCertificate(hello)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return cert, nil
		}
	}

	tlsConfig := &tls.Config{
		NextProtos:               []string{`h2`, `http/1.1`},
		GetCertificate:           getCertificate,
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	return tlsConfig, nil
}

func getDesignatedCertificate(certFile, keyFile string) (*tls.Certificate, error) {

	tlsCert, err := getCertificate(ioutil.ReadFile, certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, `could not process designated certificate`)
	}

	return tlsCert, nil
}

func getEmbeddedCertificate() (*tls.Certificate, error) {

	tlsCert, err := getCertificate(bindata.Asset, `cert.pem`, `key.pem`)
	if err != nil {
		return nil, errors.Wrap(err, `could not process embedded certificate`)
	}

	return tlsCert, nil
}

type bufferGetter func(string) ([]byte, error)

func getCertificate(getter bufferGetter, certFile, keyFile string) (*tls.Certificate, error) {

	certBuf, err := getter(certFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	keyBuf, err := getter(keyFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tlsCert, err := tls.X509KeyPair(certBuf, keyBuf)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &tlsCert, nil
}
