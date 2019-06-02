//Package security creates valid user objects for the system
// Starter code credit to:  github.com/mura123yasu/go-cognito@master:6/1/2019  (base starter)
// Updated to remove region checks - pull and check from the JWT itself instead
// Updated to add JWK cache - will enable refresh in future, I don't think AWS rotates yet
// Updated to remove redundant checks with the JWT library
// Changed the order significantly so that it can self serve in library
// Modified to be more consistent within this overall frameowrk (more work to do here later: future)
package security

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var jwks map[string]map[string]JWKKey

func init() {
	jwks = make(map[string]map[string]JWKKey)
}

//Returns JWK to use as long as it is coming from cognito service
func getJSONWebKey(issuer string) (map[string]JWKKey, error) {
	//Already in Cache?
	jwk := jwks[issuer]
	if len(jwk) != 0 {
		return jwk, nil
	}

	//Validate issuer is AWS
	parsedURL, err := url.Parse(issuer)
	if err != nil {
		return map[string]JWKKey{}, fmt.Errorf("Issuer value is not valid")
	}
	hostname := parsedURL.Hostname()
	if !strings.HasSuffix(hostname, "amazonaws.com") {
		return map[string]JWKKey{}, fmt.Errorf("Invalid JWT host, only Cognito supported so far: %v", hostname)
	}

	//Download it, we have not yet seen this issuer yet
	jwk = getJWK(issuer + "/.well-known/jwks.json")
	jwks[issuer] = jwk
	return jwk, nil
}

func validateToken(tokenStr string) (*jwt.Token, error) {

	// 2. Decode the token string into JWT format.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

		// cognito user pool : RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// pull the issuer within AWS
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("claims object not of expected type: %v", token.Claims)
		}

		// Get the JWK
		jwk, err := getJSONWebKey(claims["iss"].(string))

		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve the JWK, error: %v", err)
		}

		// 5. Get the kid from the JWT token header and retrieve the corresponding JSON Web Key that was stored
		if kid, ok := token.Header["kid"]; ok {
			if kidStr, ok := kid.(string); ok {
				key := jwk[kidStr]
				// 6. Verify the signature of the decoded JWT token.
				rsaPublicKey := convertKey(key.E, key.N)
				return rsaPublicKey, nil
			}
		}
		return nil, fmt.Errorf("Unexpected error for key id: %v", token.Header["kid"])
	})

	if err != nil {
		return token, err
	}

	if token.Valid {
		return token, nil
	}

	return token, err
}

func validateClaimItem(key string, keyShouldBe []string, claims jwt.MapClaims) error {
	if val, ok := claims[key]; ok {
		if valStr, ok := val.(string); ok {
			for _, shouldbe := range keyShouldBe {
				if valStr == shouldbe {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%v does not match any of valid values: %v", key, keyShouldBe)
}

// https://gist.github.com/MathieuMailhos/361f24316d2de29e8d41e808e0071b13
func convertKey(rawE, rawN string) *rsa.PublicKey {
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		panic(err)
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		panic(err)
	}
	pubKey.N.SetBytes(decodedN)
	// fmt.Println(decodedN)
	// fmt.Println(decodedE)
	// fmt.Printf("%#v\n", *pubKey)
	return pubKey
}

// JWK is json data struct for JSON Web Key
type JWK struct {
	Keys []JWKKey
}

// JWKKey is json data struct for cognito jwk key
type JWKKey struct {
	Alg string
	E   string
	Kid string
	Kty string
	N   string
	Use string
}

func getJWK(jwkURL string) map[string]JWKKey {

	jwk := &JWK{}

	getJSON(jwkURL, jwk)

	jwkMap := make(map[string]JWKKey, 0)
	for _, jwk := range jwk.Keys {
		jwkMap[jwk.Kid] = jwk
	}
	return jwkMap
}

func getJSON(url string, target interface{}) error {
	var myClient = &http.Client{Timeout: 3 * time.Second}
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
