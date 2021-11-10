package config

import (
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
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

// Exported globale variables
var (
	Conf            Config
	TargetURL       *url.URL
	Log             *logrus.Entry
	IdleTimeout     time.Duration
	SessionLifetime time.Duration
)

type OidcConfig struct {
	ClientID     string   `yaml:"clientID"`     // OAuth2 client ID of this application.
	ClientSecret string   `yaml:"clientSecret"` // "OAuth2 client secret of this application."
	IssuerURL    string   `yaml:"issuerURL"`    // URL of the OpenID Connect issuer.
	RedirectURL  string   `yaml:"redirectURL"`  // Callback URL for OAuth2 responses. Domain must be same as initial call, for cookies to be shared;
	Scopes       []string `yaml:"scopes"`       // The scopes we will request from the OIDC server. Default: "profile"
	RootCAFile   string   `yaml:"rootCAFile"`   // The root CA file for validation of IssuerURL
	Debug        bool     `yaml:"debug"`        // Print all request and responses from the OpenID Connect issuer.
}

type SessionConfig struct {
	IdleTimeout string `yaml:"idleTimeout"` // The maximum length of time a session can be inactive before being expired
	Lifetime    string `yaml:"lifetime"`    // The absolute maximum length of time that a session is valid.
}

type Config struct {
	configFolder   string
	LogLevel       string        `yaml:"logLevel"`       // INFO,DEBUG, ....
	LogMode        string        `yaml:"logMode"`        // Log output format: 'dev' or 'json'
	BindAddr       string        `yaml:"bindAddr"`       // The address to listen on. (default to :9001)
	TargetURL      string        `yaml:"targetURL"`      // The URL to forward all requests
	OidcConfig     OidcConfig    `yaml:"oidc"`           // OIDC client config
	Passthroughs   []string      `yaml:"passthroughs"`   // Paths pattern to forward without authentication (See http.ServeMux for path definition)
	TokenDisplay   bool          `yaml:"tokenDisplay"`   // Display an intermediate token page after login (Debugging only)
	SessionConfig  SessionConfig `yaml:"sessionConfig"`  // Web session parameters
	UserConfigFile string        `yaml:"userConfigFile"` // File hosting allowed users/groups
}
