# Dexgate, an OIDC authentication proxy. 

## Overview

`Dexgate` is a reverse proxy intended to authenticate access to a web front end.

It is intended to be used in front of a web server which does not provide access control. It is a simple gatekeeper, with a logic of 'access all' or 'no access at all'.

Designed to be used in a kubernetes context, it is intended to be inserted between the ingress controller and the web application service.

It works in cooperation with an OIDC authentication server. Currently, the only one tested is [DEX](https://github.com/dexidp/dex), one of the most used in a Kubernetes context.


## How it works

A good understanding of the interaction between the componants is required to setup a working configuration.

As depicted below, `dexgate` is inserted beetween the ingress controller of the web application and the web application itself (To its Kubernetes Service in fact).

![](docs/dexgate-Overview1.jpg)

A typical successful user interaction is the following:

1. The user intend to access the targeted Web Application. For this it send an http request targeted to the corresponding ingress controller endpoint. This ingress controller 
will forward the request to `dexgate`.
2. `dexgate` handle the request. As there is no valid session cookie in this request, a new HTTP session is initiated and the requested URL is saved in. Then `dexgate` reply 
with a redirect response targeting a login URL provided by the dex server. (Accessed throught its own ingress controller)
3. The `dex` server handle the request by sending back a Login form.
4. The user fill the form and post information to the `dex` server.
5. The `dex` server check the provided credentials against the `ldap` server. Which provide also some other information about the user (Email, groups, ...). 
The `dex` server then reply to the user by sending a redirect to the callback URL of dexgate, with a specific code as one of the parameter.
6. On receipt of this redirection `dexgate` read the code and send a request to the `dex` server to validate this authentication code.
7. The `dex` server reply with both an oauth2 token and an OIDC JWT token corresponding to the code
8. `dexgate` decode the OIDC token witch carry information about the user, such as name, email, groups, ... (Such info are name `claims` in OIDC jargon). 
Based on these informations, user access is validated.
9. `dexgate` store the oauth token in the session, to mark it as valid and redirect the user to the initial requested page, stored in step 2.
10. The user's browser resend the same request as in step 1. But, this time, there is an active and valid HTTP session, `dexgate` forward the request to the targeted web application.
11. The user can now freely access the targeted web application until the HTTP session expire. 

A more formal description of this could be find [here](docs/dexgate-Sequence.jpg)

Alternate interaction:
- If the user is not authenticated in step 5, `dex` will resend the login page.
- If the user is authenticated but its profile does not allow `dexgate` to grant access, it will be redirected to an 'allowed' page. 

### Initialisation:

An OIDC server provide a set of entry points for different action (User login interaction, code validation, token renewal, etc....). 
Fortunately, for an administrator, there is only one to define: the so called 'issuer URL'. 
When `dexgate` start, it will retrieve  the OIDC configuration from a 'well known' path, based on this issuer URL. 
For example, if the issuer URL is  `https://dex.ingress.my.cluster.com/dex`, it will send a request to `https://dex.ingress.my.cluster.com/dex/.well-known/openid-configuration`

## The Issuer URL.

A stated above, one of the main configuration parameter is the 'Issuer URL' 

- This URL must be defined both in `dex` and `dexgate` configuration with the same value.
- All other endpoints are based on this issuer URL (Same scheme and host).
- As the user login URL is based on, this URL must be reachable from outside the kubernetes cluster.




When `dexgate` is starting, it will request 

## Configuration reference

## Users authorisation Hot Reload

## login URL overriding

## Deployment


## Used library

Bsed on Dex example app

Http Session management
OIDC client
Reverse proxy
