package main

import (
	"dexgate/internal/config"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var log *logrus.Entry

// From net/http/clone.go
func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func main() {
	config.Setup()
	log = config.GetLog()
	log.Infof("Dexgate %s listening at '%s' to forward to '%s' (Logleve:%s)", config.GetVersion(), config.GetBindAddr(), config.GetTargetURL(), config.GetLogLevel())
	targetURL := config.GetTargetURL()
	targetQuery := targetURL.RawQuery
	director := func(req *http.Request) {
		oldUrl := cloneURL(req.URL)
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// Following is copied from Director of NewSingleHostReverseProxy()
		req.URL.Path, req.URL.RawPath = joinURLPath(targetURL, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		log.Debugf("%s %s -> %s", req.Method, oldUrl.String(), req.URL.String())
	}
	reverseProxy := &httputil.ReverseProxy{Director: director}

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
