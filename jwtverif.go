package jwtverif

// Copyright (C) Philip Schlump, 2009-2019.

// Verify JWT tokens - create them using CLI tool

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pschlump/dbgo"
	"github.com/pschlump/json"
)

// From: crud.go:5882(ish)

// // xyzzy-JWT
// func CreateJWTToken(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
//
// 	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, dbgo.LF())
// 	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, dbgo.LF())
//
// 	// func SignToken(tokData []byte, keyFile string) (out string, err error) {
// 	//	hdlr.KeyFilePrivate        string                      // private key file for signing JWT tokens
// 	// https://github.com/dgrijalva/jwt-go.git
//
// 	type RedirectToData struct {
// 		Status    string   `json:"status"`
// 		JWTClaims []string `json:"$JWT-claims$"`
// 	}
//
// 	var ed RedirectToData
// 	var all map[string]interface{}
//
// 	err := json.Unmarshal([]byte(rv), &ed)
// 	if err != nil {
// 		return rv, false, 200
// 	}
// 	err = json.Unmarshal([]byte(rv), &all)
// 	if err != nil {
// 		return rv, false, 200
// 	}
//
// 	if ed.Status == "success" && len(ed.JWTClaims) > 0 {
//
// 		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, dbgo.LF())
//
// 		claims := make(map[string]string)
// 		for _, vv := range ed.JWTClaims {
// 			claims[vv] = all[vv].(string)
// 			// delete(all, vv)
// 		}
// 		tokData := dbgo.SVar(claims)
//
// 		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, dbgo.LF())
//
// 		signedKey, err := SignToken([]byte(tokData), hdlr.KeyFilePrivate)
// 		if err != nil {
// 			all["status"] = "error"
// 			all["msg"] = fmt.Sprintf("Error: Unable to sign the JWT token, %s", err)
// 			delete(all, "$JWT-claims$")
// 			rv = dbgo.SVar(all)
//
// 			fmt.Printf("Error: Unable to sign the JWT token, %s\n", err)
// 			fmt.Fprintf(os.Stderr, "Error: Unable to sign the JWT token, %s\n", err)
// 			return rv, true, 406
// 		}
//
// 		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top signedKey = -->>%s<<-- %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, signedKey, dbgo.LF())
//
// 		all["jwt_token"] = signedKey
//
// 		delete(all, "$JWT-claims$")
//
// 		rv = dbgo.SVar(all)
// 		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, dbgo.LF())
// 		return rv, false, 200
// 	}
//
// 	return rv, false, 200
// }

// Create, sign, and output a token.  This is a great, simple example of
// how to use this library to create and sign a token.
func SignToken(tokData []byte, keyFile, keyType string) (out string, err error) {

	// parse the JSON of the claims
	var claims jwt.MapClaims
	if err = json.Unmarshal(tokData, &claims); err != nil {
		err = fmt.Errorf("Couldn't parse claims JSON: %v", err)
		return
	}

	fmt.Printf("Siging: %s\n", tokData)
	fmt.Printf("Claims: %s\n", dbgo.SVarI(claims))

	//-	// add command line claims
	//-	if len(flagClaims) > 0 {
	//-		for k, v := range flagClaims {
	//-			claims[k] = v
	//-		}
	//-	}

	// get the key
	var key interface{}
	key, err = loadData(keyFile)
	if err != nil {
		err = fmt.Errorf("Couldn't read key: %v", err)
		return
	}

	var Alg string
	if isEs(keyType) {
		Alg = keyType // "ES256"
	} else if isRs(keyType) {
		Alg = keyType // "RS256"
	}
	// get the signing alg
	alg := jwt.GetSigningMethod(Alg)
	if alg == nil {
		err = fmt.Errorf("Couldn't find signing method: [%v]", Alg) // xyzzy Param
		return
	}

	// create a new token
	token := jwt.NewWithClaims(alg, claims)

	//-	// add command line headers
	//-	if len(flagHead) > 0 {
	//-		for k, v := range flagHead {
	//-			token.Header[k] = v
	//-		}
	//-	}

	if isEs(keyType) {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseECPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	} else if isRs(keyType) {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseRSAPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	} else {
		err = fmt.Errorf("Error signing token - confg error: keyType=[%s]", keyType)
	}

	if out, err = token.SignedString(key); err == nil {
		if db81 {
			fmt.Println(out)
		}
	} else {
		err = fmt.Errorf("Error signing token: %v", err)
	}

	return
}

func isEs(s string) bool {
	return strings.HasPrefix(s, "ES")
}

func isRs(s string) bool {
	return strings.HasPrefix(s, "RS")
}

// Helper func:  Read input from specified file or stdin
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}

	var rdr io.Reader
	//	if p == "-" {
	//		rdr = os.Stdin
	//	} else if p == "+" {
	//		return []byte("{}"), nil
	//	} else {
	if f, err := os.Open(p); err == nil {
		rdr = f
		defer f.Close()
	} else {
		return nil, err
	}
	//	}
	return ioutil.ReadAll(rdr)
}

// Verify a token and output the claims.  This is a great example
// of how to verify and view a token.
func VerifyToken(tokData []byte, keyFile, keyType string) (iat string, err error) {

	// trim possible whitespace from token
	tokData = regexp.MustCompile(`\s*$`).ReplaceAll(tokData, []byte{})
	if db100 {
		fmt.Fprintf(os.Stderr, "Token len: %v bytes\n", len(tokData))
	}

	// Parse the token.  Load the key from command line option
	token, err := jwt.Parse(string(tokData), func(t *jwt.Token) (interface{}, error) {
		data, err := loadData(keyFile)
		if err != nil {
			return nil, err
		}
		if isEs(keyType) {
			return jwt.ParseECPublicKeyFromPEM(data)
		} else if isRs(keyType) {
			return jwt.ParseRSAPublicKeyFromPEM(data)
		} else {
			err = fmt.Errorf("Error signing token - confg error: keyType=[%s]", keyType)
			return "", err
		}
		return data, nil
	})

	// Print some debug data
	if db100 && token != nil {
		fmt.Fprintf(os.Stderr, "Header:\n%v\n", token.Header)
		fmt.Fprintf(os.Stderr, "Claims:\n%v\n", token.Claims)
	}

	// Print an error if we can't parse for some reason
	if err != nil {
		return "", fmt.Errorf("Couldn't parse token: %v", err)
	}

	// Is token invalid?
	if !token.Valid {
		return "", fmt.Errorf("Token is invalid")
	}

	if db100 {
		fmt.Fprintf(os.Stderr, "Token Claims: %s\n", dbgo.SVarI(token.Claims))
	}

	// {"auth_token":"f5d8f6ae-e2e5-42c9-83a9-dfd07825b0fc"}
	type GetAuthToken struct {
		AuthToken string `json:"auth_token"`
	}
	var gt GetAuthToken
	cl := dbgo.SVar(token.Claims)
	if db100 {
		fmt.Fprintf(os.Stderr, "Claims just before -->>%s<<--\n", cl)
	}
	err = json.Unmarshal([]byte(cl), &gt)
	if err == nil {
		if db100 {
			fmt.Fprintf(os.Stderr, "Success: %s -- token [%s] \n", err, gt.AuthToken)
		}
		return gt.AuthToken, nil
	} else {
		if db100 {
			fmt.Fprintf(os.Stderr, "Error: %s -- Unable to unmarsal -->>%s<<--\n", err, cl)
		}
		return "", err
	}

}

const db11 = false
const db81 = false
const db100 = false
