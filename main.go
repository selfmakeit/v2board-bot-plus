package main

import (
	conf "crisp_tg_bot/config"
	
	"log"

	// "time"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

var bot *tgbotapi.BotAPI
var config *viper.Viper
var redisClient *redis.Client
func init() {
	config = conf.GetConfig()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.host"),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.db"),
	})

	var err error

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Panic(err)
	}

	InitDB()

	log.Printf("Initializing Bot...")

	bot, err = tgbotapi.NewBotAPI(config.GetString("telegram.key"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.GetBool("debug")
	// tgbotapi.Remov

	log.Printf("Authorized on account %s", bot.Self.UserName)

}

func main() {

	var updates tgbotapi.UpdatesChannel

	log.Print("Start pooling")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates = bot.GetUpdatesChan(u)

	for update := range updates {

		if update.FromChat()!=nil && (update.FromChat().IsGroup() || update.FromChat().IsSuperGroup()) {
			go handleGroup(&update)
		} else if update.FromChat()!=nil && update.FromChat().IsPrivate() {
			go handlePrivate(&update)
		}

		

	}
}

func GetUserIdentifier(u tgbotapi.User) string {
	if u.UserName == "" {
		return u.FirstName
	}
	return u.UserName

}


/*

func AssemblyMsg(chatId int64, msgId int, f interface{}, page ...int) {
	var newMsg tgbotapi.EditMessageTextConfig
	fn := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	fmt.Println("----------->", fn)
	if strings.Contains(fn, "getDayTrafficStatistic") {
		if len(page) > 0 {

			var row []tgbotapi.InlineKeyboardButton
			if page[0] > 0 {
				newMsg = tgbotapi.NewEditMessageText(chatId, msgId, f.(func(int) string)(page[0]))
				btn1 := tgbotapi.NewInlineKeyboardButtonData("◀️上一页", packBtnMsg(TRAFFIC_PAGE, page[0]-1))
				btn2 := tgbotapi.NewInlineKeyboardButtonData("下一页▶️", packBtnMsg(TRAFFIC_PAGE, page[0]+1))
				row = tgbotapi.NewInlineKeyboardRow(btn1, btn2)
			} else {
				load := tgbotapi.NewMessage(chatId, "请稍等......")
				load.DisableNotification = true
				mm, _ := bot.Send(load)
				newMsg = tgbotapi.NewEditMessageText(chatId, mm.MessageID, f.(func(int) string)(page[0]))
				btn2 := tgbotapi.NewInlineKeyboardButtonData("下一页▶️", packBtnMsg(TRAFFIC_PAGE, page[0]+1))
				row = tgbotapi.NewInlineKeyboardRow(btn2)
			}
			keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
			newMsg.ReplyMarkup = &keyboard
		}
	} else {
		load := tgbotapi.NewMessage(chatId, "请稍等......")
		load.DisableNotification = true
		mm, _ := bot.Send(load)
		newMsg = tgbotapi.NewEditMessageText(chatId, mm.MessageID, f.(func() string)())
		newMsg.ParseMode = "Markdown"
	}
	bot.Send(newMsg)
	// deleteMessage(chatId, msgId, time.Millisecond*1)
}

*/
