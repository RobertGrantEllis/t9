package autocert

import (
	"net"

	"github.com/pkg/errors"
	"gopkg.in/asaskevich/govalidator.v8"
)

const (
	httpChallengeAddressDefault = `127.0.0.1:4240`
)

type Configuration struct {
	HttpChallengeAddress string   `json:"http_challenge_address"`
	ServiceAddress       string   `json:"-"`
	CacheDir             string   `json:"cache_dir"`
	HostWhitelist        []string `json:"host_whitelist"`
}

func NewConfiguration() Configuration {

	return Configuration{
		HttpChallengeAddress: httpChallengeAddressDefault,
	}
}

func (configuration *Configuration) Normalize() error {

	if len(configuration.HttpChallengeAddress) == 0 {
		configuration.HttpChallengeAddress = httpChallengeAddressDefault
	} else if _, _, err := net.SplitHostPort(configuration.HttpChallengeAddress); err != nil {
		return errors.Wrap(err, `invalid ACME HTTP challenge address`)
	}

	if len(configuration.ServiceAddress) == 0 {
		return errors.New(`service address is required`)
	} else if _, _, err := net.SplitHostPort(configuration.ServiceAddress); err != nil {
		return errors.Wrap(err, `invalid service address`)
	}

	if len(configuration.CacheDir) == 0 {
		return errors.New(`cache directory is required`)
	}

	if len(configuration.HostWhitelist) == 0 {
		return errors.New(`host whitelist is required`)
	}

	for _, host := range configuration.HostWhitelist {

		if len(host) == 0 {
			return errors.New(`host is empty`)
		} else if !govalidator.IsDNSName(host) {
			return errors.Errorf(`invalid host (received '%s')`, host)
		}
	}

	return nil
}

func (configuration *Configuration) getServicePort() int {

	_, servicePort, err := splitHostPort(configuration.ServiceAddress, 0)
	if err != nil {
		// we call this function only after normalization so this should never happen
		panic(err)
	} else if servicePort == 0 {
		// again this should never happen since we call this function only after normalization
		panic(errors.New(`port is unknown`))
	}

	return servicePort
}
