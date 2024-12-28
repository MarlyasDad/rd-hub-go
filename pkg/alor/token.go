package alor

import "time"

func NewToken() Token {
	return Token{}
}

type Token struct {
	Acccess           string
	Info              TokenInfo
	Refresh           string
	RefreshExpiration time.Time
}

type TokenInfo struct {
	Ent        string    `json:"ent"`        // value: client
	ClientId   int64     `json:"clientid"`   // value: 115177
	Portfolios []string  `json:"portfolios"` // value: 750054G D38572 G14708
	Exp        time.Time `json:"exp"`        // (Expiration Time): Время истечения value: 1.722433536e+09
	Iat        time.Time `json:"iat"`        // (Issued At): Выдано value: 1.722431736e+09
	Aud        []string  `json:"aud"`        // (Audience): Аудитория value: Client WARP WarpATConnector subscriptionsApi CommandApi InstrumentApi TradingViewPlatform Hyperion
	Sub        string    `json:"sub"`        // (Subject): Тема value: P0XXXXX
	Ein        int64     `json:"ein"`        // value: 37817
	Azp        string    `json:"azp"`        // value: 90def4c530d54f62981b
	Agreements int64     `json:"agreements"` // value: 38572
	Scope      []string  `json:"scope"`      // value: OrdersRead OrdersCreate Trades Personal Stats
	Iss        string    `json:"iss"`        // (Issuer): Издатель value: Alor.Identity
}

func (t Token) IsExpired() bool {
	return t.RefreshExpiration.Before(time.Now())
}

func (t Token) HoursToExpiration() int {
	dur := time.Since(t.RefreshExpiration)

	return int(dur.Hours())
}

func (t Token) ChangeRefresh(refresh string, expiration time.Time) {
}
