package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func addOrUpdateUser(db *sql.DB, userID int64, username, firstName, lastName string) error {
	query := `
        INSERT INTO users (user_id, username, first_name, last_name) 
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO UPDATE SET 
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name;
    `
	_, err := db.Exec(query, userID, username, firstName, lastName)
	if err != nil {
		return fmt.Errorf("ошибка записи пользователя в базу данных: %v", err)
	}
	return nil
}

func HelloUser(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, "+update.Message.From.UserName)
	bot.Send(msg)
}

func GetCurrentTime(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	t := time.Now()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, t.Format("15:04:05"))
	bot.Send(msg)
}

func GetHelp(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	helpMessage := "Доступные команды:\n" +
		"/start - Начать взаимодействие с ботом\n" +
		"/time - Получить текущее время\n" +
		"/joke - Получить случайную шутку\n" +
		"/roll - Поиграем?\n" +
		"/help - Получить список доступных команд"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
	bot.Send(msg)
}

var jokes = []string{
	"— Доктор, у меня память плохая.\n— И давно это у вас?\n— Что давно?",
	"— Вовочка, что за шум?\n— Это я линейку уронил.\n— Что ж тогда все смеются?\n— Я её на глобус уронил.",
	"– Почему вы опоздали?\n– Да так, проснулся, оделся, посмотрел на часы, а там уже два нуля.",
}

func GetJoke(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, jokes[rand.Intn(len(jokes))])
	bot.Send(msg)
}

func RollDice(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	DiceUserResult := rand.Intn(12) + 1
	DiceBotResult := rand.Intn(12) + 1

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вам выпало: "+strconv.Itoa(DiceUserResult)+" \n"+"Боту выпало: "+strconv.Itoa(DiceBotResult))

	var res string
	switch {
	case DiceUserResult > DiceBotResult:
		res = "Вы победили!!!"
	case DiceBotResult > DiceUserResult:
		res = "Бот победил!"
	default:
		res = "Ничья"
	}

	msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, res)
	bot.Send(msg)
	bot.Send(msg2)

}

func HandleCommands(update tgbotapi.Update, bot tgbotapi.BotAPI) {
	switch update.Message.Text {
	case "/start":
		HelloUser(update, bot)
	case "/time":
		GetCurrentTime(update, bot)
	case "/joke":
		GetJoke(update, bot)
	case "/roll":
		RollDice(update, bot)
	case "/help":
		GetHelp(update, bot)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
		bot.Send(msg)
	}

}

func main() {

	connStr := "user=postgres password=/// dbname=tgbot1 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Успешно подключено к базе данных PostgreSQL!")

	bot, err := tgbotapi.NewBotAPI("///")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать взаимодействие с ботом"},
		{Command: "time", Description: "Получить текущее время"},
		{Command: "joke", Description: "Шуточку?"},
		{Command: "roll", Description: "Поиграем?"},
		{Command: "help", Description: "Помочь?"},
	}

	_, err = bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			user := update.Message.From
			userID := user.ID
			username := user.UserName
			firstName := user.FirstName
			lastName := user.LastName

			err := addOrUpdateUser(db, userID, username, firstName, lastName)
			if err != nil {
				log.Println("Ошибка записи информации о пользователе:", err)
			} else {
				log.Println("Информация о пользователе успешно сохранена или обновлена.")
			}

			HandleCommands(update, *bot)
		}
	}
}
