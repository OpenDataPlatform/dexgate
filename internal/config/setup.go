package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
	var oidcDebugDefault bool = true
	var oidcDebug *bool = &oidcDebugDefault

	pflag.StringVar(&configFile, "config", "config.yml", "Configuration file")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&bindAddr, "bindAddr", ":9001", "The address to listen on.")
	pflag.StringVar(&targetUrl, "targetUrl", "", "All requests will be forwarded to this URL")
	pflag.BoolVar(oidcDebug, "oidcDebug", false, "Print all request and responses from the OpenID Connect issuer.")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	err := loadConfig(configFile, &conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to load config file: %v\n", err)
		os.Exit(2)
	}
	adjustConfigString(pflag.CommandLine, &conf.LogLevel, "logLevel")
	adjustConfigString(pflag.CommandLine, &conf.LogMode, "logMode")
	adjustConfigString(pflag.CommandLine, &conf.BindAddr, "bindAddr")
	adjustConfigString(pflag.CommandLine, &conf.TargetURL, "targetUrl")
	adjustConfigBool(pflag.CommandLine, &conf.OidcConfig.Debug, "oidcDebug")
	if conf.LogMode != "dev" && conf.LogMode != "json" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Invalid logMode value: %s. Must be one of 'dev' or 'json'\n", conf.LogMode)
		os.Exit(2)
	}
	if conf.TargetURL == "" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: TargetUrl must be defined\n")
		os.Exit(2)
	}
	conf.targetURL, err = url.Parse(conf.TargetURL)
	if err != nil || (conf.targetURL.Scheme != "http" && conf.targetURL.Scheme != "https") {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: '%s' is not a valid url\n", conf.TargetURL)
		os.Exit(2)
	}

	llevel, ok := logLevelByString[conf.LogLevel]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s is an invalid value for logLevel\n", conf.LogLevel)
		os.Exit(2)
	}
	conf.log = logrus.WithFields(logrus.Fields{})
	conf.log.Logger.SetLevel(llevel)
	if conf.LogMode == "json" {
		conf.log.Logger.SetFormatter(&logrus.JSONFormatter{})
	}
}

//
//func adjustPath(baseFolder string, path *string) {
//	if *path != "" {
//		if !filepath.IsAbs(*path) {
//			*path = filepath.Join(baseFolder, *path)
//		}
//		*path = filepath.Clean(*path)
//	}
//}
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

func adjustConfigBool(flagSet *pflag.FlagSet, inConfig **bool, param string) {
	var err error
	var ljson bool
	if flagSet.Lookup(param).Changed {
		if ljson, err = flagSet.GetBool(param); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
			os.Exit(2)
		}
		*inConfig = &ljson
	} else if *inConfig == nil {
		if ljson, err = strconv.ParseBool(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
		*inConfig = &ljson
	}
}
