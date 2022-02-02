# Dexgate, an OIDC authentication proxy. 

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Index

- [Overview](#overview)
- [How it works](#how-it-works)
  - [Alternate interaction](#alternate-interaction)
  - [Initialisation:](#initialisation)
- [Configuration](#configuration)
  - [Entry points](#entry-points)
  - [Users permissions](#users-permissions)
  - [Command line](#command-line)
  - [The Issuer URL.](#the-issuer-url)
    - [login URL overriding](#login-url-overriding)
- [Deployment](#deployment)
- [Components](#components)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

`Dexgate` is a reverse proxy intended to authenticate access to a web front end.

It is intended to be used in front of a web server which does not provide access control. It is a simple gatekeeper, with a logic of 'access all' or 'no access at all'.

Designed to be used in a kubernetes context, it is intended to be inserted between the ingress controller and the web application service.

It works in cooperation with an OIDC authentication server. Currently, the only one tested is [DEX](https://github.com/dexidp/dex), one of the most used in a Kubernetes context.

## How it works

A good understanding of the interaction between the components is required to setup a working configuration.

As depicted below, `dexgate` is inserted beetween the ingress controller of the web application and the web application itself (To its Kubernetes Service in fact).

![](docs/dexgate-Overview1.jpg)

A typical successful user interaction is the following:

1. The user intend to access the targeted Web Application. For this it send an http request targeted to the corresponding ingress controller endpoint. This ingress controller 
will forward the request to `dexgate`.
2. `dexgate` handle the request. As there is no valid session cookie in this request, a new HTTP session is initiated and the requested URL is saved in this session. Then `dexgate` reply 
with a redirect response targeting a login URL provided by the dex server. (Accessed through its own ingress controller)
3. The `dex` server handle the request by sending back a Login form to the user.
4. The user fill the form and post information to the `dex` server.
5. The `dex` server check the provided credentials against the `ldap` server. Which provide also some other informations about the user (Email, groups, ...). 
The `dex` server then reply to the user by sending a redirect to the callback URL of `dexgate`, with a specific code as one of the parameter.
6. On receipt of this redirection `dexgate` read the code and send a request to the `dex` server to validate this authentication code.
7. The `dex` server reply with both an oauth2 token and an OIDC JWT token corresponding to the code
8. `dexgate` decode the OIDC token witch carry information about the user, such as name, email, groups, ... (Such info are named `claims` in OIDC jargon). 
Based on this information, user access is validated.
9. `dexgate` store the oauth token in the session, to mark it as valid and redirect the user to the initial requested page, stored in step 2.
10. The user's browser resend the same request as in step 1. But, this time, there is an active and valid HTTP session, `dexgate` forward the request to the targeted web application.
11. The user can now freely access the targeted web application until the HTTP session expire. 

A more formal description of this could be find [here](docs/dexgate-Sequence.jpg)

### Alternate interaction

- If the user is not authenticated in step 5, `dex` will resend the login page.
- In step8, if the user is authenticated but its profile does not allow access to the target application to be granted, it will be redirected to an 'unallowed' page. 

### Initialisation:

An OIDC server provide a set of entry points for different action (User login interaction, code validation, token renewal, etc....). 

Fortunately, for an administrator, there is only one to define: the so called `issuer URL`. 

When `dexgate` start, it will retrieve  the OIDC configuration from a 'well known' path, based on this `issuer URL`.

For example, if the issuer URL is  `https://dex.ingress.my.cluster.com/dex`, it will send a request to `https://dex.ingress.my.cluster.com/dex/.well-known/openid-configuration`

This URL can also be accessed with a simple `curl`:

```
$ curl https://dex.ingress.my.cluster.com/dex/.well-known/openid-configuration
{
  "issuer": "https://dex.ingress.my.cluster.com/dex",
  "authorization_endpoint": "https://dex.ingress.my.cluster.com/dex/auth",
  "token_endpoint": "https://dex.ingress.my.cluster.com/dex/token",
  "jwks_uri": "https://dex.ingress.my.cluster.com/dex/keys",
  "userinfo_endpoint": "https://dex.ingress.my.cluster.com/dex/userinfo",
  "device_authorization_endpoint": "https://dex.ingress.my.cluster.com/dex/device/code",
  ....
```

Fortunatly, `dexgate` handle this for you. You don't have to bother with all these URL. 

## Configuration

`Dexgate` configuration is performed using two separate files: One for the general configuration (`config.yml`) and one describing users permissions (`users.yml`).

The main reason for such separation is that the configuration has no reason to change once setup is completed, while users permissions change are usual.

The default configuration file is `config.yml`. Its name and path can be overriden using the `--config` parameter.

| Name                        | req.   | Default     | Description                                                                                                                                                                                                       |
|-----------------------------|--------|-------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| logLevel                    | No     | INFO        | Log level (PANIC, FATAL, ERROR, WARN, INFO, DEBUG, TRACE)                                                                                                                                                         |
| logMode                     | No     | json        | In which form log are generated:<br>- `json`: Appropriate for further indexing.<br>- `dev`: More human readable                                                                                                   |
| bindAddr                    | No     | :9001       | The address `dexgate` will be listening on.                                                                                                                                                                       |
| targetURL                   | Yes    |             | The internal URL of the targeted web application. Typically, refer to a K8s Service                                                                                                                               |
| oidc.clientID               | No (1) |             | OAuth2 client ID of this application                                                                                                                                                                              |
| oidc.clientIDEnv            | No (1) |             | An environment variable hosting the OAuth2 client ID of this application                                                                                                                                          |
| oidc.clientSecret           | No (2) |             | The secret associated to this client ID                                                                                                                                                                           |
| oidc.clientSecretEnv        | No (2) |             | An environment variable hosting the secret associated to this client ID                                                                                                                                           |
| oidc.issuerURL              | Yes    |             | The OIDC server main URL entry. See above                                                                                                                                                                         |
| oidc.redirectURL            | Yes    |             | Where the OIDC server will redirect the user once authenticated. Must ends with `/dg_callback` (See 'entry points' below). For security reasons, this URL must be also provided in the OIDC server configuration. |
| oidc.scopes                 | No     | ["profile"] | A list of string defining the type of user information we want to grab from the user. Typically can be ["profile", "email", "groups"]                                                                             |
| oidc.rootCAFile             | No     |             | The root Certificate Authority used to validate the HTTPS exchange with the `Issuer URL` (Not needed if the `Issuer URL` is HTTP)                                                                                 |
| oidc.loginURLOverride       | No     |             | Allow override of `scheme` and `host:port` of the user login URL. See below                                                                                                                                       |
| oidc.debug                  | No     | False       | Add a bunch of message for OIDC exchange. Quite verbose. To use only for debuging                                                                                                                                 |
| passthroughs                | No     | []          | A list or URL Path which will go through `dexgate` without any authorisation. A typical usage is to set to [ "/favicon.ico" ]                                                                                     |
| tokenDisplay                | No     | False       | Display an intermediate page after login, providing tokens values and associated information. For debugging only.                                                                                                 |
| sessionConfig.idleTimeout   | No     | 15m         | The maximum time the user HTTP session can be inactive before being expired                                                                                                                                       |
| sessionConfig.lifeTime      | No     | 6h          | The absolute maximum time the user HTTP session is valid.                                                                                                                                                         |
| userConfigFile              | No (3) |             | The path (Relative to config file) providing users permissions (Exclusive from `userConfigMap.*` parameter). See 'Users permissions' below                                                                        |
| userConfigMap.configMapName | No (3) |             | The name of the Kubernetes configMap hosting the users permissions. (Exclusive from `userConfigFile` parameter). See 'Users permissions' below                                                                    |
| userConfigMap.namespace     | No     | Current ns  | The namespace of the above configMap. Default to the `dexgate`'s one.                                                                                                                                             |
| userConfigMap.configMapKey  | No     | users.yml   | The key inside the configMap hosting the users permissions yaml data.                                                                                                                                             |

(1), (2), (3): Defining one and only one of this couple of variable is required

Here is a sample of a minimalist config file:

```
---
logLevel:   "INFO"
targetURL: "http://apache2.apachenamespace.svc:80"
passthroughs: [ "/favicon.ico" ]
usersConfigFile: users.yml
oidc:
  clientID: dexgate
  clientSecret: qh8CIbdJbTYg64rtrzZ5NMg
  issuerURL: https://dex.mycluster.mycompany.com/dex
  redirectURL: https://apache.ingress.mycluster.mycompany.com/dg_callback
  rootCAFile: ca.crt
  scopes:
    - profile
    - groups
    - email
```

### Entry points

Dexgate offer several entry points:

| Name          | Usage                                                                                                                               |
|---------------|-------------------------------------------------------------------------------------------------------------------------------------|
| /dg_callback  | This is where the OIDC server will have to redirect the user on successful authentication                                           |
| /dg_unallowed | This is where `dexgate` redirect the user when not granted to access the required resource                                            |
| /dg_logout    | This URL may be called explicitly in a session to clear this current HTTP session.                                                  |
| /dg_info      | This URL may be called explicitly in a session to display user's token information. For debugging usage                             |
| /*            | All others path will be forwarded the the target site if there is an HTTP session. Otherwise, the authentication process is started |

To call `/dg_logout` and `/dg_info`, just issue a request on an URL similar to the `redirectURL` parameter by replacing `dg_callback` by `dg_logout` or `dg_info`. 

For example: 

```
`https://apache.ingress.mycluster.mycompany.com/dg_logout`
```

### Users permissions

The user permissions yaml is just made of 3 entries:

| Name          | req. | Def. | Description                 |
|---------------|----|------|-----------------------------|
| allowedUsers  | No | []   | List of allowed user names  |
| allowedGroups | No | []   | List of allowed user groups |
| allowedEmails | No | []   | List of allowed user emails |

Here is a simple sample:

```
---
allowedUsers:
- "Adam SMITH"
- "Karl MARX"
allowedGroups:
- developers
allowedEmails:
- "theboss@mycompany.com"
```

This yaml can be provided in two ways:

- As a regular yaml file, where the path is provided by the `userConfigFile` parameter.
- As a config map, as provided by the `userConfigMap.name/namespace/key` parameters

Only one of these configurations must be used.

In both case, `Dexgate` provide a 'hot reload' mechanisme, watching for any change to immediately reload the configuration. Note the following:

- Existing sessions are not impacted by such reload.
- Only this `users.yml` file/configMap is dynamicaly reloaded. Dexgate needs to be restarted to take in account any modification in this main `config.yml` file. 
- In a Kubernetes context, the usual practice would be to mount a configMap as a volume and use the `userConfigFile` parameter to point on it. But the file watcher will not work with such mount.
This is why a configMap kubernetes watcher has been implemented and the recommended pattern in kubernetes is to use the `userConfigMap.name/namespace/key` parameters.

### Command line

Also, some configuration parameters can be overridden on the command line:

```
$ ./dexgate --help
Usage of ../../dexgate/bin/dexgate:
--config string                    Configuration file (default "config.yml")
--logLevel string                  Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE) (default "INFO")
--logMode string                   Log mode: 'dev' or 'json' (default "json")
--bindAddr string                  The address to listen on. (default ":9001")
--targetUrl string                 All requests will be forwarded to this URL
--oidcDebug                        Print all request and responses from the OpenID Connect issuer.
--tokenDisplay                     Display an intermediate token page after login (Debugging only).
--idleTimeout string               The maximum length of time a session can be inactive before being expired (default "15m")
--sessionLifetime string           The absolute maximum length of time that a session is valid. (default "6h")
--oidcRootCAFile string            Root CA for validation of issuer URL.
--usersConfigFile string           Users/Groups permission file.
--usersConfigMapNamespace string   Users/Groups permission configMap namespace.
--usersConfigMapName string        Users/Groups permission configMap name.
--usersConfigMapKey string         Users/Groups permission key in configMap. (default "users.yml")
--loginURLOverride string          Allow overriding of scheme and host part of the login URL provided by the OIDC server.
```

### The Issuer URL.

A stated above, one of the main configuration parameter is the `Issuer URL` 

- This URL must be defined both in `dex` and `dexgate` configuration with the same value.
- All other endpoints are based on this issuer URL (Same scheme and host).
- As the user login URL is based on, this URL must be reachable from outside the kubernetes cluster.

This last constraint means the `issuer URL` will typically be handled by the entry load balancer/ingress controler. 
And, as there is only one configuration value, this means the connexion between `dexgate` and `dex` will also go through this path. 
So, the effective interaction is more like the following:

![](docs/dexgate-Overview2.jpg)

#### login URL overriding

There may be some network configuration where such path will not work. It may be impossible for a pod to reach another one by using the external entry point, as depicted above.

So, the solution is to define the `Issuer URL` using the kubernetes internal `dex` service address, like `http://dex.<dexnamespace>.svc:5556/dex`. 
Doing so, the communication between `dexgate` and `dex` will be direct, as in the first picture.

**But, the login URL which is also based on this adress, will be unreachable from outside the Kubernetes context.**

To fix this, a specific parameter has been added to override the URL sent to the user for the login. Add the following in the configuration:

```
oidc:
  loginURLOverride: https://dex.ingress.my.cluster.com   # Adjust the URL to your context
```

This will override the login URL. Note than only the scheme and host part (Including port) can be overriden.

## Deployment

A example Helm chart is provided in the `example` folder. This is the easiest way to deploy `dexgate`

Use this as a starting point. You may need to adjust this chart to your own needs.

Also you will need to define `dexgate` as a client in the `dex` configuration. This can be achieved by one of the following :
 
- Defining an entry in the `staticClients` list of the `dex` configuration.
- Using the dex gRPC API, which allow dynamic client creation.
- Providing `dexNamespace` and `encodedClientID` information is the `values.yaml` file. This will generate a `OAuth2Client` kubernetes resource. 
More information on this can be found [here](https://github.com/OpenDataPlatform/dexi2n)

There is also several pattern for defining clientID/Secret, straight in the `values.yaml` file, or by storing them in a kubernetes secret.
You will find more information on this as comments in the `values.yaml` file.

## Components

`Dexgate` is in fact mainly the integration of pre-existing components.

Mostly based on the [Dex example application](https://github.com/dexidp/dex/tree/master/examples/example-app), it also uses : 

- The standard [Golang reverse proxy feature](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- The standard [OAuth2 library](https://pkg.go.dev/golang.org/x/oauth2)
- The [coreos OIDC library used by `dex`](https://github.com/coreos/go-oidc)
- The [alexedwards HTTP session management library](https://github.com/alexedwards/scs)

Thanks to all of them.
