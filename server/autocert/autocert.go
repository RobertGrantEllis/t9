package autocert

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/acme/autocert"

	"github.com/RobertGrantEllis/t9/logger"
)

type Manager interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

func New(configuration Configuration, l logger.Logger) (Manager, error) {

	if err := configuration.Normalize(); err != nil {
		return nil, errors.Wrap(err, `could not instantiate Autocert`)
	}

	if err := os.MkdirAll(configuration.CacheDir, 0700); err != nil {
		return nil, errors.Wrap(err, `could not create cache directory`)
	}

	autocertManager := &autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       autocert.DirCache(configuration.CacheDir),
		RenewBefore: 24 * time.Hour,
		HostPolicy:  autocert.HostWhitelist(configuration.HostWhitelist...),
	}

	m := &manager{
		Manager:     autocertManager,
		servicePort: configuration.getServicePort(),
	}

	l.Infof(
		`starting ACME HTTP challenge listener at %s`,
		configuration.HttpChallengeAddress,
	)

	if err := m.start(configuration.HttpChallengeAddress, l); err != nil {
		return nil, err
	}

	return m, nil
}

type manager struct {
	*autocert.Manager
	servicePort int
}

func (m *manager) start(addr string, l logger.Logger) error {

	fallbackHandler := http.HandlerFunc(m.handleRedirectToHTTPS)
	handler := m.Manager.HTTPHandler(fallbackHandler)

	httpServer := &http.Server{
		Handler:  handler,
		ErrorLog: l.GetLogger(logger.WarnLevel),
	}

	listener, err := net.Listen(`tcp`, addr)
	if err != nil {
		return errors.Wrap(err, `could not start ACME HTTP challenge listener`)
	}

	errChan := make(chan error, 0)
	go func(ch chan<- error) {
		ch <- httpServer.Serve(listener)
		close(errChan)
	}(errChan)

	select {
	case err := <-errChan:
		return errors.Wrap(err, `could not start ACME HTTP challenge server`)
	case <-time.After(10 * time.Millisecond):
	}

	return nil
}

func (m *manager) handleRedirectToHTTPS(rw http.ResponseWriter, request *http.Request) {

	// strip port off host if specified
	host, _, err := splitHostPort(request.Host, 0)
	if err != nil {
		log.Printf(`error splitting host and port from http request: '%s'`, err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if m.servicePort != 443 {
		host = fmt.Sprintf(`%s:%d`, host, m.servicePort)
	}

	redirectURL := &url.URL{
		Scheme:   `https`,
		Host:     host,
		Path:     request.URL.Path,
		RawQuery: request.URL.RawQuery,
	}

	rw.Header().Set(`Location`, redirectURL.String())
	rw.WriteHeader(http.StatusFound)
}
