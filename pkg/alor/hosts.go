package alor

type (
	Hosts struct {
		Authorization, Data, Websocket string
	}

	Circuits struct {
		Production  Hosts
		Development Hosts
	}
)

var circuits = Circuits{
	Production: Hosts{
		Authorization: "https://oauth.alor.ru",
		Data:          "https://api.alor.ru",
		Websocket:     "wss://api.alor.ru/ws",
	},
	Development: Hosts{
		Authorization: "https://oauthdev.alor.ru",
		Data:          "https://apidev.alor.ru",
		Websocket:     "wss://apidev.alor.ru/ws",
	},
}
