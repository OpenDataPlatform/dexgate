package templates

import (
	"html/template"
	"net/http"
)

var logoutTmpl = template.Must(template.New("token.html").Parse(`<html>
  <head>
    <style>
/* make pre wrap */
pre {
 white-space: pre-wrap;       /* css-3 */
 white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
 white-space: -pre-wrap;      /* Opera 4-6 */
 white-space: -o-pre-wrap;    /* Opera 7 */
 word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
    </style>
  </head>
  <body>
	<p>You are now logged out</p>
	<p>Bye!</p>
	<input type="button" onclick="location.href='{{ .LandingURL }}';" value="Re-enter.... ">
  </body>
</html>
`))

type logoutTmplData struct {
	LandingURL string
}

func RenderLogout(w http.ResponseWriter, landingURL string) {
	renderTemplate(w, logoutTmpl, logoutTmplData{
		LandingURL: landingURL,
	})
}
