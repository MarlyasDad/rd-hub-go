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

func NewAuthorization(host string, client http.Client, token Token) Authorization {
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
	a.client.Timeout = time.Duration(2 * time.Second)

	url := fmt.Sprintf("%s/refresh?token=%s", a.Host, a.Token.Refresh)

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// Submit the request
	res, err := a.client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		log.Println("Forbidden!")
		return errors.New("broker answered forbidden")
	}

	body, _ := io.ReadAll(res.Body)

	var r RefreshResponse

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println(err)
		return err
	}

	a.Token.Acccess = r.AccessToken

	claims := jwt.MapClaims{}

	// Headers map[alg:ES256 typ:JWT] without kid
	_, _, err = jwt.NewParser().ParseUnverified(a.Token.Acccess, claims)
	if err != nil {
		log.Println(err)

		if err == jwt.ErrSignatureInvalid {
			log.Println("Invalid Token Signature")
			return err
		}
		return err
	}

	// for key, val := range claims {
	// 	fmt.Printf("Key: %v, value: %v\n", key, val)
	// }

	tokenInfo, err := NewTokenInfo(claims)
	if err != nil {
		log.Println(err)
		return err
	}

	a.Token.Info = tokenInfo

	return nil
}

func NewTokenInfo(claims jwt.MapClaims) (TokenInfo, error) {
	var info TokenInfo

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
