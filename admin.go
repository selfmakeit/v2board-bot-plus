package main

import (
	"strconv"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processAdminCommand(update *tgbotapi.Update) {
	var msg tgbotapi.MessageConfig
	upmsg := update.Message
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
	if config.Get("isAutoDeleteMsg").(bool) {

		msgToDelete := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, upmsg.MessageID)
		_, _ = bot.Request(msgToDelete)
	}
	switch upmsg.Command() {
	case "start", "help":
		msg.Text = fmt.Sprintf("🤖V2board机器人\n\n当前会话id:`%v`\n\n你的账户id:`%v`", update.Message.Chat.ID, update.Message.From.ID)
		btn1 := tgbotapi.NewInlineKeyboardButtonData("🧰查看菜单", packBtnMsg(SHOW_MENU, SHOW_MENU))
		row := tgbotapi.NewInlineKeyboardRow(btn1)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
		msg.ReplyMarkup = keyboard
		// btn2 := tgbotapi.NewInlineKeyboardButtonData("查询数据2", "query2")
		msg.ParseMode = "Markdown"
		sendMessage(msg)
	case "shop":
	case "ticket":
	case "myinvite":
	}
}
func processAdminBtnCallBack(update *tgbotapi.Update) {
	if update.CallbackQuery != nil {
		// 获取按钮的回调数据
		data, err := unPackBtnMsg(update.CallbackQuery.Data)
		if err != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error())
			bot.Send(msg)
		}
		// 根据不同的回调数据进行不同的操作
		switch data.Type {

		case HIDE_USER_INFO:
			msgToDelete := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			_, _ = bot.Request(msgToDelete)
		case SHOW_MENU:
			bt1 := tgbotapi.NewKeyboardButton("💰最近收益")
			bt2 := tgbotapi.NewKeyboardButton("📈用户增长")
			bt3 := tgbotapi.NewKeyboardButton("💌邀请统计")
			bt4 := tgbotapi.NewKeyboardButton("🛒套餐分析")
			bt5 := tgbotapi.NewKeyboardButton("📊今日流量排行")
			bt6 := tgbotapi.NewKeyboardButton("📊本月流量排行")
			row1 := tgbotapi.NewKeyboardButtonRow(bt1, bt2)
			row2 := tgbotapi.NewKeyboardButtonRow(bt3, bt4)
			row3 := tgbotapi.NewKeyboardButtonRow(bt5)
			row4 := tgbotapi.NewKeyboardButtonRow(bt6)
			keyboard := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
			keyboard.ResizeKeyboard = true
			// keyboard.OneTimeKeyboard = true
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "请选择")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case TRAFFIC_PAGE: //翻页
			p, err := strconv.Atoi(data.Value.(string))
			if err == nil {
				AssemblyMsg(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, getDayTrafficStatistic, TRAFFIC_PAGE,p, 1)
			}
		case TRAFFIC_M_PAGE: //翻页
			p, err := strconv.Atoi(data.Value.(string))
			if err == nil {
				AssemblyMsg(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, getMonthTrafficStatistic, TRAFFIC_M_PAGE,p, 1)
			}
		case INVITE_PAGE: //翻页
			p, err := strconv.Atoi(data.Value.(string))
			if err == nil {
				AssemblyMsg(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, getInviteStatistic, INVITE_PAGE,p, 1)
			}
		case PLAN_PAGE: //翻页
			p, err := strconv.Atoi(data.Value.(string))
			if err == nil {
				AssemblyMsg(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, getPlanStatistic, PLAN_PAGE,p, 1)
			}
		}

		// 回复按钮点击事件，使按钮的选中状态消失,不然会弹按钮里data的消息提示
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		bot.Send(callback)
		var v string
		if data.Value == nil {
			v = ""
		} else {
			v = data.Value.(string)
		}
		bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, v))
	}
}

func processAdminTxt(update *tgbotapi.Update) {
	if update.Message.Text == "💰最近收益" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID, getIncomeStatistic,"")
	} else if update.Message.Text == "📈用户增长" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID, getIncrementStatistic,"")
	} else if update.Message.Text == "💌邀请统计" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID, getInviteStatistic,INVITE_PAGE,0)
	} else if update.Message.Text == "🛒套餐分析" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID,  getPlanStatistic, PLAN_PAGE,0)
	} else if update.Message.Text == "📊今日流量排行" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID, getDayTrafficStatistic, TRAFFIC_PAGE,0)
	} else if update.Message.Text == "📊本月流量排行" {
		AssemblyMsg(update.Message.Chat.ID, update.Message.MessageID, getMonthTrafficStatistic,TRAFFIC_M_PAGE, 0)
	} else {
		processReply(update)
	}
}
func processReply(update *tgbotapi.Update) {
	if update.Message.ReplyToMessage == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请选择一条消息回复")
		bot.Send(msg)
		return
	}
	res, err := redisClient.Get(strconv.Itoa(update.Message.ReplyToMessage.MessageID)+"forward").Result()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR: "+err.Error())
		bot.Send(msg)
		return
	}
	chatId, _ := strconv.ParseInt(res, 10, 64)
	msg := tgbotapi.NewCopyMessage(chatId, update.Message.Chat.ID, update.Message.MessageID)
	bot.Send(msg)
}

// extra 第一个参数是页码，第二是标记是否是第一次查看流量排行,第二个随便传，是根据长度来判断的
func AssemblyMsg(chatId int64, msgId int, f interface{},pt string, extra ...int) {

	var newMsg tgbotapi.EditMessageTextConfig
	fn := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	if strings.Contains(fn, "getDayTrafficStatistic") || strings.Contains(fn, "getMonthTrafficStatistic")|| strings.Contains(fn, "getInviteStatistic") || strings.Contains(fn, "getPlanStatistic") {
		if len(extra) > 0 {
			var row []tgbotapi.InlineKeyboardButton
			if len(extra) > 1 {
				newMsg = tgbotapi.NewEditMessageText(chatId, msgId, f.(func(int) string)(extra[0]))
			} else {
				load := tgbotapi.NewMessage(chatId, "请稍等......")
				load.DisableNotification = true
				mm, _ := bot.Send(load)
				newMsg = tgbotapi.NewEditMessageText(chatId, mm.MessageID, f.(func(int) string)(extra[0]))
			}
			if extra[0] > 0 {
				btn1 := tgbotapi.NewInlineKeyboardButtonData("◀️上一页", packBtnMsg(pt, extra[0]-1))
				btn2 := tgbotapi.NewInlineKeyboardButtonData("下一页▶️", packBtnMsg(pt, extra[0]+1))
				row = tgbotapi.NewInlineKeyboardRow(btn1, btn2)

			} else {
				btn2 := tgbotapi.NewInlineKeyboardButtonData("下一页▶️", packBtnMsg(pt, extra[0]+1))
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
	newMsg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(newMsg)
	// deleteMessage(chatId, msgId, time.Millisecond*1)
}
