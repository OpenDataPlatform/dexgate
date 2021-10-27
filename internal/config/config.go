package config

import (
	"github.com/sirupsen/logrus"
	"net/url"
)

var logLevelByString = map[string]logrus.Level{
	"PANIC": logrus.PanicLevel,
	"FATAL": logrus.FatalLevel,
	"ERROR": logrus.ErrorLevel,
	"WARN":  logrus.WarnLevel,
	"INFO":  logrus.InfoLevel,
	"DEBUG": logrus.DebugLevel,
	"TRACE": logrus.TraceLevel,
}

// Conf Global variable
var conf = Config{}

type OidcConfig struct {
	ClientID     string `yaml:"clientID"`     // OAuth2 client ID of this application.
	ClientSecret string `yaml:"clientSecret"` // "OAuth2 client secret of this application."
	IssuerURL    string `yaml:"issuerURL"`    // URL of the OpenID Connect issuer.
	RedirectURL  string `yaml:"redirectURL"`  // Callback URL for OAuth2 responses.

}

type Config struct {
	configFolder string
	LogLevel     string     `yaml:"logLevel"`  // DEBUG, ....
	LogMode      string     `yaml:"logMode"`   // Log output format: 'dev' or 'json'
	BindAddr     string     `yaml:"bindAddr"`  // The address to listen on. (default to :9001)
	TargetURL    string     `yaml:"targetURL"` // The URL to forward all requests
	OidcConfig   OidcConfig `yaml:"oidc"`      // OIDC client config
	// Transformed data
	targetURL *url.URL
	log       *logrus.Entry
}

func GetLog() *logrus.Entry {
	return conf.log
}

func GetTargetURL() *url.URL {
	return conf.targetURL
}

func GetBindAddr() string {
	return conf.BindAddr
}

func GetVersion() string {
	return version
}

func GetLogLevel() string {
	return conf.LogLevel
}

func GetOidcConfig() *OidcConfig {
	return &conf.OidcConfig
}
