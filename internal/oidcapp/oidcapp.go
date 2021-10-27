package oidcapp

import (
	"context"
	"dexgate/internal/config"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"net/http"
)

type oidcApp struct {
	config         *config.OidcConfig
	client         *http.Client
	provider       *oidc.Provider
	verifier       *oidc.IDTokenVerifier
	offlineAsScope bool
}

func NewOidcApp(config *config.OidcConfig, debug bool) (*oidcApp, error) {
	var err error
	app := &oidcApp{
		config: config,
	}
	// We build a specific http.client, for
	// - Allowing some Debug on exchange
	// - Setup SSL connection (TODO)
	if debug {
		if app.client == nil {
			app.client = &http.Client{
				Transport: debugTransport{http.DefaultTransport},
			}
		} else {
			app.client.Transport = debugTransport{app.client.Transport}
		}
	}
	if app.client == nil {
		app.client = http.DefaultClient
	}
	ctx := oidc.ClientContext(context.Background(), app.client)
	app.provider, err = oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider %q: %v", config.IssuerURL, err)
	}
	app.verifier = app.provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	// Following is copied from dex/exammple/example.-app/main.go
	var s struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := app.provider.Claims(&s); err != nil {
		return nil, fmt.Errorf("failed to parse provider scopes_supported: %v", err)
	}
	if len(s.ScopesSupported) == 0 {
		// scopes_supported is a "RECOMMENDED" discovery claim, not a required
		// one. If missing, assume that the provider follows the spec and has
		// an "offline_access" scope.
		app.offlineAsScope = true
	} else {
		// See if scopes_supported has the "offline_access" scope.
		app.offlineAsScope = func() bool {
			for _, scope := range s.ScopesSupported {
				if scope == oidc.ScopeOfflineAccess {
					return true
				}
			}
			return false
		}()
	}
	return app, nil
}

func (app *oidcApp) oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     app.config.ClientID,
		ClientSecret: app.config.ClientSecret,
		Endpoint:     app.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  app.config.RedirectURL,
	}
}

const dexgateAppState = "WhatANiceAuthStuff"

func (app *oidcApp) NewLoginURL() string {
	scopes := []string{"openid", "profile", "email", "groups"}
	if app.offlineAsScope {
		scopes = append(scopes, "offline_access")
		return app.oauth2Config(scopes).AuthCodeURL(dexgateAppState)
	} else {
		return app.oauth2Config(scopes).AuthCodeURL(dexgateAppState, oauth2.AccessTypeOffline)
	}
}
