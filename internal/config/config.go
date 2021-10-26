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

type Config struct {
	configFolder string
	LogLevel     string `yaml:"logLevel"`  // DEBUG, ....
	LogMode      string `yaml:"logMode"`   // Log output format: 'dev' or 'json'
	BindAddr     string `yaml:"bindAddr"`  // The address to listen on. (default to :9001)
	TargetUrl    string `yaml:"targetUrl"` // The URL to forward all requests
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
