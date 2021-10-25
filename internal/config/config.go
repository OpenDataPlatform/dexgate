package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
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

type Config struct {
	configFolder string
	LogLevel     string `yaml:"logLevel"`  // DEBUG, ....
	LogMode      string `yaml:"logMode"`   // Log output format: 'dev' or 'json'
	BindAddr     string `yaml:"bindAddr"`  // The address to listen on. (default to :9001)
	TargetUrl    string `yaml:"targetUrl"` // The URL to forward all requests
	// Transformed data
	targetURL *url.URL
}

func GetLog() *logrus.Entry {
	llevel, ok := logLevelByString[conf.LogLevel]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s is an invalid value for logLevel\n", conf.LogLevel)
		os.Exit(2)
	}
	log := logrus.WithFields(logrus.Fields{})
	log.Logger.SetLevel(llevel)
	if conf.LogMode == "json" {
		log.Logger.SetFormatter(&logrus.JSONFormatter{})
	}
	return log
}

func GetTargetURL() *url.URL {
	return conf.targetURL
}

func GetBindAddr() string {
	return conf.BindAddr
}
