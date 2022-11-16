package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	owm "github.com/briandowns/openweathermap"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var InLineMenu = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Погода в данный момент", Currentpogoda("Фрязино"))),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Прогноз на 5 дней", "Не работает")),
)

func Currentpogoda(city string) string {
	w, err := owm.NewCurrent("C", "ru", "WeatherOpenMap_Token")
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByName(city)
	if w.Name != "" {
		return "Сейчас " + strconv.FormatFloat(math.Round(w.Main.Temp), 'f', 0, 64) + "°C " + w.Weather[0].Description
	}
	return "Некорректный город"
}

func gorod(city string) bool {
	w, err := owm.NewCurrent("C", "ru", "WeatherOpenMap_Token")
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByName(city)
	if w.Name != "" {
		return true
	}
	return false
}

func TelegramBot(db *sql.DB) {

	bot, err := tgbotapi.NewBotAPI("TelegramBot_Token")
	if err != nil {
		log.Panic(err)
	}
	var id int
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			row := db.QueryRow("select count(*) from User where User_ID = $1", update.Message.Chat.ID)
			row.Scan(&id)
			if id == 0 {
				switch update.Message.Text {
				case "/start":

					//Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, я погодный бот. Введи свой город чтобы всегда знать погоду в нём")
					bot.Send(msg)

				default:
					if gorod(update.Message.Text) {

						result, err := db.Exec("insert into user (User_ID, User_City) values ($1, $2)", update.Message.Chat.ID, update.Message.Text)
						if err != nil {
							panic(err)
						}
						fmt.Println(result.RowsAffected())
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я запомнил ваш город")
						msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Mеню")))
						bot.Send(msg)

					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой город")
						bot.Send(msg)
					}
				}
			} else {
				switch update.Message.Text {
				case "Mеню":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
					msg.ReplyMarkup = InLineMenu
					bot.Send(msg)
				}

			}
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
			// And finally, send a message containing the data received.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func main() {

	db, err := sql.Open("sqlite3", "E:/GoRepos/sql/GoDB.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	TelegramBot(db)

}
