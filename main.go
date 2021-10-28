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
- Logout
- Page _info_
- Make session duratiopn parameter configurable
- A list of URL passthroughs (i.e: /favicon.ico)
- implements allowed email/groups lists
- Enable SSL on input
- Enable SSL/CA on client
- Remove user approval for scope (Dex config ?)
- Perform retry to allow late Dex startup
- Intégration kube
- Documentation
*/
var log *logrus.Entry

var tokenCount int = 0

func newToken() string {
	tokenCount++
	return fmt.Sprintf("%05d", tokenCount)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("LOGOUT"))
}

func dumpHeader(r *http.Request) {
	for name, values := range r.Header {
		for _, value := range values {
			log.Debugf("%s %s", name, value)
		}
	}
}

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
	mux.Handle("/callback", callbackHandler(oidcApp))
	mux.Handle("/", mainHandler(sessionManager, reverseProxy, oidcApp))
	log.Fatal(http.ListenAndServe(config.GetBindAddr(), sessionManager.LoadAndSave(mux)))
}

func mainHandler(sessionManager *scs.SessionManager, reverseProxy *httputil.ReverseProxy, oidcApp *oidcapp.OidcApp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := sessionManager.GetString(r.Context(), "token")
		if token == "" {
			// Fresh session. Must enter login process
			lurl := oidcApp.NewLoginURL()
			log.Debugf("%s %s => Not logged. Will redirect to %s", r.Method, r.URL, lurl)
			http.Redirect(w, r, lurl, http.StatusSeeOther)
			//sessionManager.Put(r.Context(), "token", token)
		} else {
			log.Debugf("%s %s => Forward to target")
			reverseProxy.ServeHTTP(w, r)
		}
	})
}

func callbackHandler(oidcApp *oidcapp.OidcApp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, errMsg := oidcApp.CheckCallbackRequest(r)
		if errMsg != "" {
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		tokenData, errMsg := oidcApp.HandleCallbackRequest(r, code)
		if errMsg != "" {
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		log.Infof("claims:%v", tokenData.Claims)

		//buff := new(bytes.Buffer)
		//if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		//	http.Error(w, fmt.Sprintf("error indenting ID token claims: %v", err), http.StatusInternalServerError)
		//	return
		//}

		templates.RenderToken(w, *tokenData)
	})

}
