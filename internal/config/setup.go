package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func loadConfig(fileName string, config *Config) error {
	configFile, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(&config); err != nil {
		return err
	}
	// All file path should be relative to the config file location. So take note of its absolute path
	config.configFolder = filepath.Dir(configFile)
	return nil
}

func Setup() {
	// Allow overriding of some config variable. Mostly used in development stage
	var configFile string
	var logLevel string
	var logMode string
	var bindAddr string
	var targetUrl string
	var oidcDebug bool
	var tokenDisplay bool
	var idleTimeout string
	var sessionLifetime string
	var userConfigFile string

	pflag.StringVar(&configFile, "config", "config.yml", "Configuration file")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&bindAddr, "bindAddr", ":9001", "The address to listen on.")
	pflag.StringVar(&targetUrl, "targetUrl", "", "All requests will be forwarded to this URL")
	pflag.BoolVar(&oidcDebug, "oidcDebug", false, "Print all request and responses from the OpenID Connect issuer.")
	pflag.BoolVar(&tokenDisplay, "tokenDisplay", false, "Display an intermediate token page after login (Debugging only).")
	pflag.StringVar(&idleTimeout, "idleTimeout", "15m", "The maximum length of time a session can be inactive before being expired")
	pflag.StringVar(&sessionLifetime, "sessionLifetime", "6h", "The absolute maximum length of time that a session is valid.")
	pflag.StringVar(&userConfigFile, "userConfigFile", "users.yml", "Users / Groups permission file.")

	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	err := loadConfig(configFile, &Conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to load config file: %v\n", err)
		os.Exit(2)
	}
	adjustConfigString(pflag.CommandLine, &Conf.LogLevel, "logLevel")
	adjustConfigString(pflag.CommandLine, &Conf.LogMode, "logMode")
	adjustConfigString(pflag.CommandLine, &Conf.BindAddr, "bindAddr")
	adjustConfigString(pflag.CommandLine, &Conf.TargetURL, "targetUrl")
	adjustConfigBool(pflag.CommandLine, &Conf.OidcConfig.Debug, "oidcDebug")
	adjustConfigBool(pflag.CommandLine, &Conf.TokenDisplay, "tokenDisplay")
	adjustConfigString(pflag.CommandLine, &Conf.SessionConfig.IdleTimeout, "idleTimeout")
	adjustConfigString(pflag.CommandLine, &Conf.SessionConfig.Lifetime, "sessionLifetime")
	adjustConfigString(pflag.CommandLine, &Conf.UserConfigFile, "userConfigFile")

	// -----------------------------------Handle logging  stuff
	if Conf.LogMode != "dev" && Conf.LogMode != "json" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Invalid logMode value: %s. Must be one of 'dev' or 'json'\n", Conf.LogMode)
		os.Exit(2)
	}
	llevel, ok := logLevelByString[Conf.LogLevel]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s is an invalid value for logLevel\n", Conf.LogLevel)
		os.Exit(2)
	}
	Log = logrus.WithFields(logrus.Fields{})
	Log.Logger.SetLevel(llevel)
	if Conf.LogMode == "json" {
		Log.Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// ------------------------------ TargetURL handling
	if Conf.TargetURL == "" {
		missingParameter("targetURL")
	}

	TargetURL, err = url.Parse(Conf.TargetURL)
	if err != nil || (TargetURL.Scheme != "http" && TargetURL.Scheme != "https") {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: '%s' is not a valid url\n", Conf.TargetURL)
		os.Exit(2)
	}
	// ------------------------- Handle Oidc config stuff
	if Conf.OidcConfig.ClientID == "" {
		missingParameter("oidcConfig.clientID")
	}
	if Conf.OidcConfig.ClientSecret == "" {
		missingParameter("oidcConfig.clientSecret")
	}
	if Conf.OidcConfig.IssuerURL == "" {
		missingParameter("oidcConfig.issuerURL")
	}
	if Conf.OidcConfig.RedirectURL == "" {
		missingParameter("oidcConfig.redirectURL")
	}

	// ----------------------- Session handling
	IdleTimeout, err = time.ParseDuration(Conf.SessionConfig.IdleTimeout)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: '%s' is not a valid Duration for 'sessionConfig.idleTimeout' parameter\n", Conf.SessionConfig.IdleTimeout)
		os.Exit(2)
	}
	SessionLifetime, err = time.ParseDuration(Conf.SessionConfig.Lifetime)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: '%s' is not a valid Duration for 'sessionConfig.lifetime' parameter\n", Conf.SessionConfig.Lifetime)
		os.Exit(2)
	}
	if Conf.UserConfigFile == "" {
		missingParameter("userConfigFile")
	}
	// Set relative to main config file
	adjustPath(Conf.configFolder, &Conf.UserConfigFile)
}

func missingParameter(param string) {
	_, _ = fmt.Fprintf(os.Stderr, "ERROR: '%s' parameter must be defined in config file\n", param)
	os.Exit(2)
}

func adjustPath(baseFolder string, path *string) {
	if *path != "" {
		if !filepath.IsAbs(*path) {
			*path = filepath.Join(baseFolder, *path)
		}
		*path = filepath.Clean(*path)
	}
}

// For all adjustConfigXxx(), we:
// - panic when error is internal
// - Display a message and exit(2) when error is from usage

func adjustConfigString(flagSet *pflag.FlagSet, inConfig *string, param string) {
	if pflag.Lookup(param).Changed {
		var err error
		if *inConfig, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
	} else if *inConfig == "" {
		*inConfig = flagSet.Lookup(param).DefValue
	}
}

//
//func adjustConfigInt(flagSet *pflag.FlagSet, inConfig *int, param string) {
//	var err error
//	if flagSet.Lookup(param).Changed {
//		if *inConfig, err = flagSet.GetInt(param); err != nil {
//			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
//			os.Exit(2)
//		}
//	} else if *inConfig == 0 {
//		if *inConfig, err = strconv.Atoi(flagSet.Lookup(param).DefValue); err != nil {
//			panic(err)
//		}
//	}
//}

//func adjustConfigBool(flagSet *pflag.FlagSet, inConfig **bool, param string) {
//	var err error
//	var ljson bool
//	if flagSet.Lookup(param).Changed {
//		if ljson, err = flagSet.GetBool(param); err != nil {
//			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
//			os.Exit(2)
//		}
//		*inConfig = &ljson
//	} else if *inConfig == nil {
//		if ljson, err = strconv.ParseBool(flagSet.Lookup(param).DefValue); err != nil {
//			panic(err)
//		}
//		*inConfig = &ljson
//	}
//}

func adjustConfigBool(flagSet *pflag.FlagSet, inConfig *bool, param string) {
	var err error
	var ljson bool
	if flagSet.Lookup(param).Changed {
		if ljson, err = flagSet.GetBool(param); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
			os.Exit(2)
		}
		*inConfig = ljson
	}
}
