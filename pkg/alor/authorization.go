package alor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewAuthorization(
	host string,
	client http.Client,
	refresh string,
	refreshExpiration time.Time,
) Authorization {

	token := Token{
		Refresh:           refresh,
		RefreshExpiration: refreshExpiration,
	}

	return Authorization{
		Host:   host,
		Token:  token,
		client: client,
	}
}

type Authorization struct {
	Host  string
	Token Token
	// TokenInfo TokenInfo
	client http.Client
}

type RefreshResponse struct {
	AccessToken string `json:"AccessToken"`
}

func (a *Authorization) Refresh() error {
	a.client.Timeout = 2 * time.Second

	url := fmt.Sprintf("%s/refresh?token=%s", a.Host, a.Token.Refresh)

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// Submit the request
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusForbidden {
		return errors.New("broker answered forbidden")
	}

	body, _ := io.ReadAll(res.Body)

	var r RefreshResponse

	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}

	a.Token.Acccess = r.AccessToken

	tokenInfo, err := a.ParseTokenInfo()
	if err != nil {
		return err
	}

	a.Token.Info = tokenInfo

	return nil
}

func extractClaims(tokenStr string) (jwt.MapClaims, bool) {
	log.Println(tokenStr)

	hmacSecretString := "30461f17-d85c-4b04-877d-d915235f0763"
	hmacSecret := []byte(hmacSecretString)
	// token, err := jwt.Parse(tokenStr, nil, jwt.WithoutClaimsValidation())
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return hmacSecret, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		log.Println(err)
		return nil, false
	}
	log.Println(token.Valid)
	// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, true
	} else {
		log.Printf("Invalid JWT Token")
		return nil, false
	}
}

func (a *Authorization) ParseTokenInfo() (TokenInfo, error) {
	var (
		info   TokenInfo
		claims jwt.MapClaims
	)

	// Headers map[alg:ES256 typ:JWT] without kid
	_, _, err := jwt.NewParser().ParseUnverified(a.Token.Acccess, &claims)
	// _, err := jwt.ParseWithClaims(a.Token.Acccess, claims, nil, jwt.WithoutClaimsValidation())
	if err != nil {

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Println("Invalid Token Signature")
			return TokenInfo{}, err
		}

		log.Println(err)
		return TokenInfo{}, err
	}

	// claims, ok := extractClaims(a.Token.Acccess)

	for key, val := range claims {
		fmt.Printf("Key: %v, value: %v\n", key, val)
	}

	ent, ok := claims["ent"].(string)
	if !ok {
		return info, errors.New("token payload: ent undefined")
	}

	clientId, err := parseClaimsInt("clientid", claims)
	if err != nil {
		return info, err
	}

	portfolios, ok := claims["portfolios"].(string)
	if !ok {
		return info, errors.New("token payload: portfolios undefined")
	}

	agreements, err := parseClaimsInt("agreements", claims)
	if err != nil {
		return info, err
	}

	ein, err := parseClaimsInt("ein", claims)
	if err != nil {
		return info, err
	}

	scope, ok := claims["scope"].(string)
	if !ok {
		return info, errors.New("token payload: scope undefined")
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return info, errors.New("token payload: iss undefined")
	}

	aud, ok := claims["aud"].(string)
	if !ok {
		return info, errors.New("token payload: aud undefined")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return info, errors.New("token payload: sub undefined")
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return info, err
	}

	iat, err := claims.GetIssuedAt()
	if err != nil {
		return info, err
	}

	azp, ok := claims["azp"].(string)
	if !ok {
		return info, errors.New("token payload: azp undefined")
	}

	return TokenInfo{
		Ent:        ent,
		ClientId:   *clientId,
		Portfolios: strings.Split(" ", portfolios),
		Exp:        exp.Time,
		Iat:        iat.Time,
		Aud:        strings.Split(" ", aud),
		Sub:        sub,
		Ein:        *ein,
		Azp:        azp,
		Agreements: *agreements,
		Scope:      strings.Split(" ", scope),
		Iss:        iss,
	}, nil
}

func parseClaimsInt(key string, m jwt.MapClaims) (*int64, error) {
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("token payload: %s undefined", key)
	}

	switch exp := v.(type) {
	case int64:
		return &exp, nil

	case json.Number:
		intVal, _ := exp.Int64()

		return &intVal, nil

	case string:
		intVal, err := strconv.Atoi(exp)

		intVal64 := int64(intVal)

		if err == nil {
			return &intVal64, nil
		}
	}

	return nil, fmt.Errorf("parse error: value %s is not int64", key)
}

func (a *Authorization) AccessToken() (string, error) {
	// check expiration time
	// if token expired call Refresh() func

	return a.Token.Acccess, nil
}
