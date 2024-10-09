package alor

func (c *AlorConnector) GetPortfolios() []string {
	return c.Authorization.Token.Info.Portfolios
}

// alor api v2 clients section
// /md/v2/Clients/:exchange/:portfolio/positions
func (c *AlorConnector) GetPortfolio(exchange string, portfolio string, format string, withoutCurrency string) {
	// /md/v2/Clients/:exchange/:portfolio/positions
	// path
	// Possible values: [MOEX, SPBX]
	// query
	// format string
	// Possible values: [Simple, Slim, Heavy]
	// withoutCurrency boolean
	//
	// Responses 200, 401, 403
}

// alor api v2 clients section
// /md/v2/Clients/:exchange/:portfolio/positions
func (c *AlorConnector) GetPortfolioPositions(exchange string, portfolio string, format string, withoutCurrency string) {
	// /md/v2/Clients/:exchange/:portfolio/positions
	// path
	// Possible values: [MOEX, SPBX]
	// query
	// format string
	// Possible values: [Simple, Slim, Heavy]
	// withoutCurrency boolean
	//
	// Responses 200, 401, 403
}
