package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/RobertGrantEllis/t9/client"
	"github.com/RobertGrantEllis/t9/server"
)

func getClientConfigurationFromArgs(args ...string) *client.Configuration {

	if len(args) == 0 {
		fail(errors.New(`no arguments provided`))
	}

	// creates a flagset and a configuration and sets up the flagset so it parses into the configuration
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configuration := client.NewConfiguration()

	fs.StringVar(&configuration.Address, `address`, configuration.Address, `address for reaching t9 server`)
	fs.DurationVar(&configuration.ConnectionTimeout, `connection-timeout`, configuration.ConnectionTimeout, `timeout for connecting to server`)
	fs.DurationVar(&configuration.RequestTimeout, `request-timeout`, configuration.RequestTimeout, `timeout for making requsts against server`)

	addHelpAndParse(fs, args...)

	return &configuration
}

type arrayFlag []string

func (af *arrayFlag) String() string {
	return strings.Join(*af, `,`)
}

func (af *arrayFlag) Set(value string) error {
	*af = append(*af, value)
	return nil
}

func getServerConfigurationFromArgs(args ...string) *server.Configuration {

	if len(args) == 0 {
		fail(errors.New(`no arguments provided`))
	}

	// creates a flagset and a configuration and sets up the flagset so it parses into the configuration
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configuration := server.NewConfiguration()

	hostWhitelist := arrayFlag{}

	fs.StringVar(&configuration.LogLevel, `log-level`, configuration.LogLevel, `debug|info|warn|error`)
	fs.StringVar(&configuration.Address, `address`, configuration.Address, `listening address for t9 server`)
	fs.StringVar(&configuration.DictionaryFile, `dictionary`, configuration.DictionaryFile, `dictionary (defaults to built-in English dictionary)`)
	fs.IntVar(&configuration.CacheSize, `cache-size`, configuration.CacheSize, `cache size for t9 words`)
	fs.StringVar(&configuration.CertificateFile, `certificate`, configuration.CertificateFile, `SSL/TLS certificate in PEM format`)
	fs.StringVar(&configuration.KeyFile, `key`, configuration.KeyFile, `SSL/TLS private key in PEM format`)
	fs.BoolVar(&configuration.UseAutocert, `use-autocert`, configuration.UseAutocert, `use autocert for TLS`)
	fs.StringVar(&configuration.Autocert.HttpChallengeAddress, `autocert-http-challenge-address`, configuration.Autocert.HttpChallengeAddress, `listening address for autocert HTTP challenges`)
	fs.StringVar(&configuration.Autocert.CacheDir, `autocert-cache-dir`, configuration.Autocert.CacheDir, `directory for storing cached certificates and keys`)
	fs.Var(&hostWhitelist, `autocert-host-whitelist`, `hosts to be whitelisted for autocert`)

	configuration.Autocert.HostWhitelist = hostWhitelist

	addHelpAndParse(fs, args...)

	return &configuration
}

func addHelpAndParse(fs *flag.FlagSet, args ...string) {

	// put in a help flag
	help := false
	fs.BoolVar(&help, `help`, help, `show this help message`)

	// suppress failure output so we can do our own thing
	fs.SetOutput(ioutil.Discard)

	// parse the arguments into the configuration
	err := fs.Parse(args[1:])

	// now we are done parsing, re-enable output writer
	fs.SetOutput(os.Stderr)

	// figure out if we want to print help
	if err != nil && err != flag.ErrHelp {
		printHelp(fs, err)
	} else if fs.NArg() != 0 {
		printHelp(fs, errors.Errorf(`unexpected argument: '%s'`, fs.Arg(0)))
	} else if help || err == flag.ErrHelp {
		printHelp(fs, nil)
	}
}

func printHelp(fs *flag.FlagSet, err error) {

	exitCode := 0
	if err != nil {
		printError(err)
		fmt.Fprintln(os.Stderr)
		exitCode = 1
	}

	fmt.Fprintf(os.Stderr, "Usage :\n")
	fs.PrintDefaults()

	os.Exit(exitCode)
}
