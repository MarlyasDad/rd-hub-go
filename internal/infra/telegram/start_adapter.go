package telegram

//func NewStartAdapter(service botDomain.ServiceStart) th.Handler {
//	return func(bot *telego.Bot, update telego.Update) {
//
//		// call service
//		response, _ := service.Start(update.Message.From.ID,
//			update.Message.From.FirstName,
//			update.Message.From.LastName,
//			update.Message.From.Username,
//			update.Message.From.LanguageCode,
//		)
//		// if err != nil {
//		// 	return
//		// 	// Send error Не Шмогла я
//		// }
//
//		// send request
//		_, _ = bot.SendMessage(&telego.SendMessageParams{
//			ChatID:    tu.ID(update.Message.Chat.ID),
//			Text:      response,
//			ParseMode: telego.ModeMarkdown,
//		})
//	}
//}
