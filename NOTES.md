# Links

https://pkg.go.dev/net/http/httputil#ReverseProxy
https://gist.github.com/JalfResi/6287706
https://www.integralist.co.uk/posts/golang-reverse-proxy/

https://medium.com/the-curious-noob/auth-reverse-proxy-in-golang-84c00b089e3a

https://github.com/Uninett/goidc-proxy
https://golangrepo.com/repo/oauth2-proxy-oauth2-proxy-go-authentication-oauth
https://github.com/oauth2-proxy/oauth2-proxy

https://www.youtube.com/watch?v=-TzLER2fX84

## Session mgmt

https://github.com/alexedwards/scs

# Tricks

tcpdump -i lo0 -A port 9999


# Spark History server

To bind to localhost: export SPARK_LOCAL_IP=localhost
See 
- https://github.com/apache/spark/blob/master/core/src/main/scala/org/apache/spark/deploy/history/HistoryServer.scala
- https://github.com/apache/spark/blob/master/core/src/main/scala/org/apache/spark/ui/WebUI.scala (Line 143)
