package main

import (
	"dexgate/internal/config"
	"dexgate/internal/director"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

var log *logrus.Entry

func main() {
	config.Setup()
	log = config.GetLog()
	log.Infof("Dexgate %s listening at '%s' to forward to '%s' (Logleve:%s)", config.GetVersion(), config.GetBindAddr(), config.GetTargetURL(), config.GetLogLevel())

	reverseProxy := &httputil.ReverseProxy{Director: director.NewDirector(config.GetTargetURL())}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//for name, values := range r.Header {
		//	for _, value := range values {
		//		log.Debugf("%s %s", name, value)
		//	}
		//}
		reverseProxy.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(config.GetBindAddr(), nil))
}
