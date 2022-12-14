package token

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

// SecretFunc is the definition of the required secret retrieval function.
type SecretFunc func(*Token) (string, error)

const (
	// UserToken is the name of the user token.
	UserToken = "user"

	// SessToken is the name of the session token.
	SessToken = "sess"

	// HookToken is the name of the hook token.
	HookToken = "hook"

	// CsrfToken is the name of the CSRF token.
	CsrfToken = "csrf"
)

// SignerAlgo defines the default algorithm used to sign JWT tokens.
const SignerAlgo = "HS256"

// Token represents our simple JWT.
type Token struct {
	Kind string
	Text string
}

// parse parses a raw JWT.
func parse(raw string, fn SecretFunc) (*Token, error) {
	token := &Token{}
	parsed, err := jwt.Parse(raw, keyFunc(token, fn))
	if err != nil {
		return nil, err
	} else if !parsed.Valid {
		return nil, jwt.ValidationError{}
	}
	return token, nil
}

// ParseRequest parses a JWT from the request.
func ParseRequest(r *http.Request, fn SecretFunc) (*Token, error) {
	var token = r.Header.Get("Authorization")

	// first we attempt to get the token from the
	// authorization header.
	if len(token) != 0 {
		token = r.Header.Get("Authorization")
		fmt.Sscanf(token, "Bearer %s", &token)
		return parse(token, fn)
	}

	// then we attempt to get the token from the
	// access_token url query parameter
	token = r.FormValue("access_token")
	if len(token) != 0 {
		return parse(token, fn)
	}

	// and finally we attempt to get the token from
	// the user session cookie
	cookie, err := r.Cookie("user_sess")
	if err != nil {
		return nil, err
	}
	return parse(cookie.Value, fn)
}

// CheckCsrf checks the validity of the JWT.
func CheckCsrf(r *http.Request, fn SecretFunc) error {

	// get and options requests are always
	// enabled, without CSRF checks.
	switch r.Method {
	case "GET", "OPTIONS":
		return nil
	}

	// parse the raw CSRF token value and validate
	raw := r.Header.Get("X-CSRF-TOKEN")
	_, err := parse(raw, fn)
	return err
}

// New initializes a new JWT.
func New(kind, text string) *Token {
	return &Token{Kind: kind, Text: text}
}

// Sign signs the token using the given secret hash
// and returns the string value.
func (t *Token) Sign(secret string) (string, error) {
	return t.SignExpires(secret, 0)
}

// SignExpires signs the token using the given secret hash
// with an expiration date.
func (t *Token) SignExpires(secret string, exp int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	mapClaims := make(jwt.MapClaims)
	mapClaims["type"] = t.Kind
	mapClaims["text"] = t.Text
	if exp > 0 {
		mapClaims["exp"] = float64(exp)
	}
	token.Claims = mapClaims
	return token.SignedString([]byte(secret))
}

func keyFunc(token *Token, fn SecretFunc) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		// validate the correct algorithm is being used
		if t.Method.Alg() != SignerAlgo {
			return nil, jwt.ErrSignatureInvalid
		}

		mapClaims, isMapClaims := t.Claims.(jwt.MapClaims)
		if !isMapClaims {
			return nil, fmt.Errorf("token claims are not stored as map")
		}

		// extract the token kind and cast to
		// the expected type.
		kindv, ok := mapClaims["type"]
		if !ok {
			return nil, jwt.ValidationError{}
		}
		token.Kind, _ = kindv.(string)

		// extract the token value and cast to
		// exepected type.
		textv, ok := mapClaims["text"]
		if !ok {
			return nil, jwt.ValidationError{}
		}
		token.Text, _ = textv.(string)

		// invoke the callback function to retrieve
		// the secret key used to verify
		secret, err := fn(token)
		return []byte(secret), err
	}
}
