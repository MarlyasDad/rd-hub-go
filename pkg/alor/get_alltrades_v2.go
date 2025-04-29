package alor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type GetAllTradesV2Params struct {
	Exchange                               Exchange
	Symbol                                 string
	Board                                  string
	JsonResponse                           bool
	From, To, FromID, ToID, QtyFrom, QtyTo *int64
	PriceFrom, PriceTo                     *float64
	Side                                   *OrderSide
	Offset                                 int64
	Take                                   int64
	Descending                             bool
	IncludeVirtualTrades                   *bool
}

func (c *Client) GetAllTrades(params GetAllTradesV2Params) ([]AllTradesSlimData, error) {
	var data []AllTradesSlimData

	// Ставим жёстко slim формат
	format := SlimResponseFormat

	// GET https://apidev.alor.ru/md/v2/Securities/:exchange/:symbol/alltrades
	method := "GET"
	url := fmt.Sprintf("%s/md/v2/Securities/%s/%s/alltrades", c.Hosts.Data, params.Exchange, params.Symbol)

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*30)
	defer cncl()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	req.Close = true

	// query parameters
	q := req.URL.Query()
	q.Add("instrumentGroup", params.Board)
	q.Add("format", string(format))
	q.Add("descending", strconv.FormatBool(params.Descending))
	q.Add("jsonResponse", strconv.FormatBool(params.JsonResponse))
	q.Add("offset", strconv.FormatInt(params.Offset, 10))

	if params.Take == 0 {
		q.Add("take", "5000")
	} else {
		q.Add("take", strconv.FormatInt(params.Take, 10))
	}

	if params.From != nil {
		q.Add("from", strconv.FormatInt(*params.From, 10))
	}

	if params.To != nil {
		q.Add("to", strconv.FormatInt(*params.To, 10))
	}

	if params.FromID != nil {
		q.Add("fromID", strconv.FormatInt(*params.FromID, 10))
	}

	if params.ToID != nil {
		q.Add("toID", strconv.FormatInt(*params.ToID, 10))
	}

	if params.QtyFrom != nil {
		q.Add("qtyFrom", strconv.FormatInt(*params.QtyFrom, 10))
	}

	if params.QtyTo != nil {
		q.Add("qtyTo", strconv.FormatInt(*params.QtyTo, 10))
	}

	if params.PriceFrom != nil {
		q.Add("priceFrom", strconv.FormatFloat(*params.PriceFrom, 'g', -1, 64))
	}

	if params.PriceTo != nil {
		q.Add("priceTo", strconv.FormatFloat(*params.PriceTo, 'g', -1, 64))
	}

	if params.Side != nil {
		q.Add("side", string(*params.Side))
	}

	if params.IncludeVirtualTrades != nil {
		q.Add("includeVirtualTrades", strconv.FormatBool(*params.IncludeVirtualTrades))
	}

	req.URL.RawQuery = q.Encode()

	accessToken, err := c.Token.GetAccessToken()
	if err != nil {
		return data, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}

	return data, nil
}
