// Package auth provides a pluggable set of "Authenticators". These Authenticators
// represent a different database used to store authorized token/tag combinations
// that, when enabled, will allow/deny access to mist methods for these authorized
// token/tags combinations.
package auth

import (
	"fmt"
	"net/url"
	"sync"
)

var (
	defaultAuth  Authenticator // defaultAuth is the current authenticator for the package; this is set during an authenticator start
	isConfigured bool          // isConfigured is used to avoid race conditions when setting or checking if defaultAuth is nil

	ErrTokenNotFound = fmt.Errorf("Token not found\n")
	ErrTokenExist    = fmt.Errorf("Token already exists\n")

	// the list of available authenticators
	authenticators = map[string]handleFunc{}
	authTex        sync.RWMutex
)

type (
	handleFunc func(url *url.URL) (Authenticator, error)

	// Authenticator represnets a database of authorized token/tag combinations.
	// These combinations are used as a way to allow access to mist methods for a
	// particular token/tag combination (when authentication is desired)
	Authenticator interface {
		AddToken(token string) error                    // add a token to list of authorized tokens
		RemoveToken(token string) error                 // remove a token from the list of authorized tokens
		AddTags(token string, tags []string) error      // add authorized tags to a token
		RemoveTags(token string, tags []string) error   // remove authorized tags from a token
		GetTagsForToken(token string) ([]string, error) // get the authorized tags for a token
	}
)

// Register registers a new mist authenticator
func Register(name string, auth handleFunc) {
	authTex.Lock()
	authenticators[name] = auth
	authTex.Unlock()
}

// IsConfigured returns whether or not an authenticator is configured and is authenticating.
func IsConfigured() bool {
	return isConfigured
}

// Start attempts to start a mist authenticator from the list of available
// authenticators; the authenticator provided is in the uri string format
// (scheme:[//[user:pass@]host[:port]][/]path[?query][#fragment])
func Start(uri string) error {

	// no authenticator is wanted
	if uri == "" {
		isConfigured = false
		return nil
	}

	// parse the uri string into a url object
	url, err := url.Parse(uri)
	if err != nil {
		return err
	}

	// check to see if the scheme is supported; if not, indicate as such and continue
	authTex.Lock()
	auth, ok := authenticators[url.Scheme]
	authTex.Unlock()
	if !ok {
		return fmt.Errorf("Unsupported scheme '%s'", url.Scheme)
	}

	// set defaultAuth by attempting to start the desired authenticator
	defaultAuth, err = auth(url)
	if err != nil {
		return err
	}
	isConfigured = true

	return nil
}
