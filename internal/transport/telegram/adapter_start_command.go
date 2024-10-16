package telegram

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type (
	HelloRequest struct {
		// json fields
		Text string
	}
)

func NewHelloAdapter(service Service) th.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		var request *HelloRequest

		request, err := getHelloRequestData(update.Message.Text)
		if err != nil {
			return
		}

		// call service
		response, err := service.EchoAnswer(request.Text)
		if err != nil {
			return
			// Send error Не Шмогла я
		}

		// send request
		_, _ = bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			fmt.Sprintf("Hello %s! %s", update.Message.From.FirstName, response),
		))
	}
}

func getHelloRequestData(data string) (*HelloRequest, error) {
	var request HelloRequest

	request.Text = data
	// or marshalling json

	return &request, nil
}
