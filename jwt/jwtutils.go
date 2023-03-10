package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type BlockwareClaims struct {
	jwt.RegisteredClaims
	AuthID   string    `json:"auth_id"`
	AuthType string    `json:"auth_type"`
	Contexts []Context `json:"contexts"`
	Purpose  string    `json:"purpose"`
	Scopes   []string  `json:"scopes"`
	Type     string    `json:"type"`
}
type Context struct {
	Handle string   `json:"handle"`
	ID     string   `json:"id"`
	Scopes []string `json:"scopes"`
	Type   string   `json:"type"`
}

func HasScopeForHandle(c echo.Context, handle string, scope string) bool {
	// Get the 'user' from the context, the user is set by the JWT middleware and is a *jwt.Token
	user := c.Get("user")
	if user == nil {
		return false
	}
	token := user.(*jwt.Token)

	if !token.Valid {
		return false
	}
	return validateScopes(token, handle, scope)

}
func validateScopes(token *jwt.Token, handle string, scope string) bool {
	// Get the scopes for the handle
	scopes := getScopesForHandle(token, handle)

	// Check if the scope is in the list of scopes
	for _, s := range scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

func getScopesForHandle(token *jwt.Token, handle string) []string {
	// Get the claims
	claims := token.Claims.(*BlockwareClaims)
	// Get the contexts
	contexts := claims.Contexts
	// Loop through the contexts
	for _, ctx := range contexts {
		// Check if the handle matches
		if ctx.Handle == handle {
			// Return the scopes
			return ctx.Scopes
		}
	}
	// Return an empty list of scopes
	return []string{}
}