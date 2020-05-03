// Shared structures and common routines
package chefapi_lib

import (

	"github.com/go-chef/chef"
	"github.com/MarkGibbons/chefapi_client"
	"github.com/dgrijalva/jwt-go"
	"errors"
	"net/http"
	"regexp"
	"strings"

)

var jwtKey = []byte("my_secret_key") // TODO: Parameter for production

// Auth is the structure returned to indicate access allowed or not
type Auth struct {
	Auth  bool   `json:"auth"`
	Group string `json:"group"`
	Node  string `json:"node"`
	Org   string `json:"org"`
	User  string `json:"user"`
}

type Claims struct {
        Username string `json:"username"`
        jwt.StandardClaims
}

// AllOrgs returns a list of all of the organizations
func AllOrgs() (orgNames []string, err error) {
        client := chefapi_client.Client()
        orgList, err := client.Organizations.List()
        if err != nil {
                return
        }
        orgNames = make([]string, 0, len(orgList))
        for k := range orgList {
                orgNames = append(orgNames, k)
        }
        return
}

// StdMessageStatus extracts the status code from a go-chef api error message
func ChefStatus(in_err error) (message string, statusCode int) {

	if in_err == nil {
		return
	}

	cerr, _ := chef.ChefError(in_err)
	if cerr != nil {
		message = cerr.Error()
		statusCode = cerr.StatusCode()
		return
	}

	message = in_err.Error()
	return
}

// Verify the input characters are allowed
func CleanInput(vars map[string]string) (err error) {
        for _, value := range vars {
                matched, _ := regexp.MatchString("^[[:word:]]+$", value)
                if !matched {
                        err = errors.New("Invalid value in the URI")
                        break
                }
        }
        return
}

// InputError sets an error message for invalid values in the url
func InputError(w *http.ResponseWriter) {
        (*w).WriteHeader(http.StatusBadRequest)
        (*w).Write([]byte(`{"message":"Bad url input value"}`))
}

// LoggedIn verifies the JWT and extracts the user name
func LoggedIn(r *http.Request) (user string, code int) {
        code = -1
        reqToken := r.Header.Get("Authorization")
        splitToken := strings.Split(reqToken, "Bearer")
        // Verify index before using
        if len(splitToken) != 2 {
                code = http.StatusBadRequest
                return
        }
        tknStr := strings.TrimSpace(splitToken[1])
        claims := &Claims{}
        tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
                return jwtKey, nil
        })
        if err != nil {
                if err == jwt.ErrSignatureInvalid {
                        code = http.StatusUnauthorized
                        return
                }
                code = http.StatusBadRequest
                return
        }
        if !tkn.Valid {
                code = http.StatusUnauthorized
                return
        }
        user = claims.Username
        return
}
