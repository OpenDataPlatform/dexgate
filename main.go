package main

import (
	"dexgate/internal/config"
	"dexgate/internal/director"
	"dexgate/internal/oidcapp"
	"dexgate/internal/templates"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"os"
)

/*
TODO
- Make token page optionnal.
- Page /dexg_logout
- Page /dexg_info_
- Rename /callback in /dexg_callback
- Make session duration parameter configurable
- implements allowed email/groups lists
- Enable SSL on input
- Enable SSL/CA on client
- Remove user approval for scope (Dex config ?)
- Perform retry to allow late Dex startup
- Intégration kube
- Documentation
*/
var log *logrus.Entry

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("LOGOUT"))
}

//func dumpHeader(r *http.Request) {
//	for name, values := range r.Header {
//		for _, value := range values {
//			log.Debugf("%s %s", name, value)
//		}
//	}
//}

func main() {
	config.Setup()
	log = config.GetLog()
	log.Infof("Dexgate %s listening at '%s' to forward to '%s' (Logleve:%s)", config.GetVersion(), config.GetBindAddr(), config.GetTargetURL(), config.GetLogLevel())

	sessionManager := scs.New()

	reverseProxy := &httputil.ReverseProxy{Director: director.NewDirector(config.GetTargetURL())}

	oidcApp, err := oidcapp.NewOidcApp(config.GetOidcConfig())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to instanciate OIDC subsystem:%v'\n", err)
		os.Exit(2)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/logout", handleLogout)
	mux.Handle("/callback", callbackHandler(sessionManager, oidcApp))
	for _, path := range config.GetPassthroughs() {
		log.Infof("Will set passthrough for %s", path)
		mux.Handle(path, passthroughHandler(reverseProxy))
	}
	mux.Handle("/", mainHandler(sessionManager, reverseProxy, oidcApp))
	log.Fatal(http.ListenAndServe(config.GetBindAddr(), sessionManager.LoadAndSave(mux)))
}

func passthroughHandler(reverseProxy *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%s %s => Forward to target (passthrough)", r.Method, r.URL)
		reverseProxy.ServeHTTP(w, r)
	})
}

/*
 We store the token and the claim in the session, as markers for logged user.
 But, we don't handle token expiration nor renewal. We rely on the session lifecycle instead
*/

func mainHandler(sessionManager *scs.SessionManager, reverseProxy *httputil.ReverseProxy, oidcApp *oidcapp.OidcApp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := sessionManager.GetString(r.Context(), "token")
		if token == "" {
			// Fresh session. Must enter login process
			lurl := oidcApp.NewLoginURL()
			log.Debugf("%s %s => Not logged. Will redirect to %s", r.Method, r.URL, lurl)
			sessionManager.Put(r.Context(), "landingURL", r.URL.String())
			http.Redirect(w, r, lurl, http.StatusSeeOther)
		} else {
			log.Debugf("%s %s => Forward to target (Authenticated)", r.Method, r.URL)
			reverseProxy.ServeHTTP(w, r)
		}
	})
}

func callbackHandler(sessionManager *scs.SessionManager, oidcApp *oidcapp.OidcApp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, errMsg := oidcApp.CheckCallbackRequest(r)
		if errMsg != "" {
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		tokenData, errMsg := oidcApp.HandleCallbackRequest(r, code)
		if errMsg != "" {
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		//log.Infof("claims:%v", tokenData.Claims)
		sessionManager.Put(r.Context(), "token", tokenData.AccessToken)
		sessionManager.Put(r.Context(), "claim", tokenData.Claims)
		landingURL := sessionManager.Get(r.Context(), "landingURL").(string)
		log.Debugf("Displaying token page")
		templates.RenderToken(w, *tokenData, landingURL)
	})
}
