package alor

func (c *Client) GetPortfolios() []string {
	//if c.Token != nil {
	//	return []string{}
	//}
	return c.Token.Data.Portfolios
}
