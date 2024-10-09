package alor

type (
	AlorHosts struct {
		Authorization, Data, Websocket string
	}

	AlorCircuits struct {
		Production  AlorHosts
		Development AlorHosts
	}
)

var circuits = AlorCircuits{
	Production: AlorHosts{
		Authorization: "https://oauth.alor.ru",
		Data:          "https://api.alor.ru",
		Websocket:     "wss://api.alor.ru/ws",
	},
	Development: AlorHosts{
		Authorization: "https://oauthdev.alor.ru",
		Data:          "https://apidev.alor.ru",
		Websocket:     "wss://apidev.alor.ru/ws",
	},
}
