package jwt

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kapetacom/insight-api/scopes"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestHasScopeForHandle(t *testing.T) {
	user := generateToken()
	fmt.Println(user)
	token, err := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{"HS256"})).ParseWithClaims(user, &KapetaClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, token)
	hasAccess := validateScopes(token, "kapeta", "*")
	assert.True(t, hasAccess)
	hasAccess = validateScopes(token, "kapeta", "random")
	assert.True(t, hasAccess)

	hasAccess = validateScopes(token, "sorenmat_org", "*")
	assert.False(t, hasAccess)
	hasAccess = validateScopes(token, "sorenmat_org", scopes.LOGGING_READ_SCOPE)
	assert.True(t, hasAccess)

}

func generateToken() string {
	// Create claims while leaving out some of the optional fields
	payload := `{
			"iss": "https://auth.kapeta.com",
			"auth_id": "63f4681a5cd2424f153ac791",
			"auth_type": "urn:ietf:params:oauth:grant-type:device_code",
			"contexts": [
			  {
				"handle": "kapeta",
				"id": "2e3d1573-07f5-4888-85d2-b6ad074b40a3",
				"scopes": [
				  "*"
				],
				"type": "organization"
			  },
			  {
				"handle": "sorenmat_org",
				"id": "db908e86-36b9-4157-b67b-ab5b13c53cdf",
				"scopes": [
				  "` + scopes.LOGGING_READ_SCOPE + `"
				],
				"type": "organization"
			  }
			],
			"exp": 1678375908,
			"iat": 1678372308,
			"iss": "https://auth.kapeta.com",
			"purpose": "access_token",
			"scopes": [
			  "offline",
			  "*"
			],
			"sub": "1a2bb1be-3624-45e9-b15d-d6fb7081fb00",
			"type": "user"
		  }`
	claims := KapetaClaims{}
	err := json.Unmarshal([]byte(payload), &claims)
	if err != nil {
		panic(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		panic(err)
	}
	return signed
}
