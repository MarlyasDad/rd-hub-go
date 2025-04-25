package alor

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetAllSecurities(ticker string, limit int64, offset int64, sector string) {
	method := "GET"
	url := "https://apidev.alor.ru/md/v2/Securities"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	accessToken, err := c.Token.GetAccessToken()
	if err != nil {
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := c.Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
