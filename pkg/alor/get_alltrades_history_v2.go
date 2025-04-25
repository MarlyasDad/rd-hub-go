package alor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (c *Client) GetAllTradesHistory(exchange string, symbol string, board string, jsonResponse bool, from, to int64, qtyFrom, qtyTo int32, priceFrom, priceTo float64, side OrderSide, offset int32, take int32, descending bool, includeVirtualTrades bool) ([]AllTradesSlimData, error) {
	var data []AllTradesSlimData

	// Ставим жёстко slim формат
	format := SlimResponseFormat

	// GET https://apidev.alor.ru/md/v2/Securities/:exchange/:symbol/alltrades
	method := "GET"
	url := fmt.Sprintf("%s/md/v2/Securities/%s/%s/alltrades/history", c.Hosts.Data, exchange, symbol)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	// from, to, qtyFrom, qtyTo, priceFrom, priceTo, side, offset, take, descending, includeVirtualTrades
	q := req.URL.Query()
	q.Add("instrumentGroup", board)
	q.Add("jsonResponse", strconv.FormatBool(jsonResponse))
	q.Add("format", string(format))
	q.Add("from", strconv.FormatInt(from, 10))
	q.Add("to", strconv.FormatInt(to, 10))

	req.URL.RawQuery = q.Encode()

	accessToken, err := c.Token.GetAccessToken()
	if err != nil {
		return data, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := c.Client.Do(req)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return data, err
	}
	// fmt.Println(string(body))

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}

	return data, nil
}
