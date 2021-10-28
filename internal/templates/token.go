package templates

import (
	"bytes"
	"dexgate/internal/oidcapp"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
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
    <p> ID Token: <pre><code>{{ .TokenData.IDToken }}</code></pre></p>
    <p> Access Token: <pre><code>{{ .TokenData.AccessToken }}</code></pre></p>
    <p> Claims: <pre><code>{{ .Claims }}</code></pre></p>
	{{ if .TokenData.RefreshToken }}
    <p> Refresh Token: <pre><code>{{ .TokenData.RefreshToken }}</code></pre></p>
	{{ end }}
	<input type="button" onclick="location.href='{{ .LandingURL }}';" value="CONTINUE....">
  </body>
</html>
`))

type tokenTmplData struct {
	TokenData  oidcapp.TokenData
	Claims     string
	LandingURL string
}

func RenderToken(w http.ResponseWriter, tokenData oidcapp.TokenData, landingURL string) {
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(tokenData.Claims), "", "  "); err != nil {
		http.Error(w, fmt.Sprintf("error indenting ID token claims: %v", err), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, tokenTmpl, tokenTmplData{
		TokenData:  tokenData,
		Claims:     buff.String(),
		LandingURL: landingURL,
	})
}
