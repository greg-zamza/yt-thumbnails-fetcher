package main

import (
    "os"
    "log"
    "strconv"

    "github.com/NicoNex/echotron/v3"
)

var bot_token = os.Getenv("BOT_TOKEN")
var password = os.Getenv("PASSWORD")

func main() {
    api := echotron.NewAPI(bot_token)

	for update := range echotron.PollingUpdates(bot_token) {
        /* эта проверка помогает боту не сломаться, если он получит
           неожиданный апдейт, который не получится обработать */
        if update.Message == nil {
            log.Println("Unhandled update")
        } else {
            //запрос в бд по ID
            //if exists {TODO}
            if update.Message.From.ID == int64(2003265450) {
                // валидация (сообщение должно содержать int < 100)
                n, err := strconv.Atoi(update.Message.Text)
                if err != nil || n > 100 || n < 1 {
                    api.SendMessage("Please send number from 1 to 100", update.ChatID(), nil)
                } else {
                    //MAIN FUNCTIONALITY
                    api.SendMessage("OKAY LEGO", update.ChatID(), nil)
                }
            } else {
                if update.Message.Text == password {
                    api.SendMessage("Welcome! 👋", update.ChatID(), nil)
                    //TODO insert to admin database
                } else {
                    api.SendMessage("please enter the password", update.ChatID(), nil)
                }
            }
        }
	}
}
