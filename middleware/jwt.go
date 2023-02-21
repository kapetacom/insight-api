package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

func Restricted() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the 'user' from the context, the user is set by the JWT middleware and is a *jwt.Token
			user := c.Get("user")
			if user == nil {
				return echo.ErrUnauthorized
			}
			token := user.(*jwt.Token)

			if !token.Valid {
				return echo.ErrUnauthorized
			}
			c.Set("jwt", token)
			return next(c)
		}
	}
}

// FetchKey is a function that returns a jwt.Keyfunc that can be used to verify the JWT
func FetchKey(url string) jwt.Keyfunc {

	return func(token *jwt.Token) (interface{}, error) {
		// Note: We download the keyset every time the restricted route is accessed.
		keySet, err := jwk.Fetch(context.Background(), url)
		if err != nil {
			log.Println("Unable to fetch the keyset. Error: ", err.Error())
			return nil, err
		}

		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("expecting JWT header to have a key ID in the kid field")
		}

		key, found := keySet.LookupKeyID(keyID)

		if !found {
			return nil, fmt.Errorf("unable to find key %q", keyID)
		}

		var pubkey interface{}
		if err := key.Raw(&pubkey); err != nil {
			return nil, fmt.Errorf("Unable to get the public key. Error: %s", err.Error())
		}
		return pubkey, nil
	}
}
