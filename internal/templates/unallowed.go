package templates

import (
	"html/template"
	"net/http"
)

var unallowedTmpl = template.Must(template.New("unallowed.html").Parse(`<html>
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
	<h2>Unallowed !</h2>
	<p>Your are not allowed to access this ressource.</p>
	<p>Refer to your system administrator</p>
  </body>
</html>
`))

type unallowedTmplData struct {
	LandingURL string
}

func RenderUnallowed(w http.ResponseWriter, landingURL string) {
	renderTemplate(w, unallowedTmpl, tokenTmplData{
		LandingURL: landingURL,
	})
}
