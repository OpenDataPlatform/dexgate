package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var infoTmpl = template.Must(template.New("token.html").Parse(`<html>
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
    <p> Access Token: <pre><code>{{ .AccessToken }}</code></pre></p>
    <p> Claims: <pre><code>{{ .Claims }}</code></pre></p>
  </body>
</html>
`))

type infoTmplData struct {
	AccessToken string
	Claims      string
}

func RenderInfo(w http.ResponseWriter, accessToken string, claims string) {
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		http.Error(w, fmt.Sprintf("error indenting ID token claims: %v", err), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, infoTmpl, infoTmplData{
		AccessToken: accessToken,
		Claims:      buff.String(),
	})
}
