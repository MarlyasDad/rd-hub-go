package alor

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type MdV2TimeResponse struct {
	// data
}

func (c *Client) GetUnixTimestamp() (int64, error) {
	path := "/md/v2/time"
	url := c.Hosts.Data + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	token, err := c.Authorization.AccessToken()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.Client.Do(req)
	if err != nil {
		// log.Println(err)
		return 0, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)

	timestamp, err := strconv.ParseInt(string(body), 16, 8)

	return timestamp, nil
}
