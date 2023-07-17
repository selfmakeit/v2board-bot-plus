package main

import (
	"errors"
	"strings"
	"time"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleGroup(update *tgbotapi.Update) {
	if update.Message != nil {
		if !IsCurrentUserAdmin(update.Message.From.ID) { //普通用户
			if update.Message.IsCommand() {
				processUserCommand(update)
				//处理文字
			} else if update.Message.Text != "" {
				processUserTxt(update)
			}
		} else { //管理员
			if update.Message.IsCommand() {
				processAdminCommand(update)
				//处理文字
			} else if update.Message.Text != "" {
				processAdminTxt(update)
			}
		}
		//处理命令

	}
	//处理按钮
	if update.CallbackQuery != nil {
		if !IsCurrentUserAdmin(update.CallbackQuery.From.ID) {
			//处理普通用户按钮
		} else {
			//处理管理员按钮
			go processAdminBtnCallBack(update)
		}
	}
	if update.ChatJoinRequest != nil {

	}
}

func handlePrivate(update *tgbotapi.Update) {
	if update.Message != nil {
		if !IsCurrentUserAdmin(update.Message.From.ID) { //普通用户
			if update.Message.IsCommand() {
				processUserCommand(update)
				//处理文字
			} else if update.Message.Text != "" {
				processUserTxt(update)
			}
		} else { //管理员
			if update.Message.IsCommand() {
				processAdminCommand(update)
				//处理文字
			} else if update.Message.Text != "" {
				processAdminTxt(update)
			}
		}
		//处理命令

	}
	//处理按钮
	if update.CallbackQuery != nil {
		if !IsCurrentUserAdmin(update.CallbackQuery.From.ID) {
			//处理普通用户按钮
			processUserBtnCallBack(update)
		} else {
			//处理管理员按钮
			processAdminBtnCallBack(update)
		}
	}
}



/**
 * 发送消息
 */
func sendMessage(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	if msg.Text == "" {
		return tgbotapi.Message{}, errors.New("txt nil")
	}
	mmsg, err := bot.Send(msg)
	if err != nil {
		return mmsg, err
	}
	// go deleteMessage(msg.ChatID, mmsg.MessageID, time.Duration(config.GetInt("redis.cacheTime"))*time.Hour)
	return mmsg, nil
}
func sendEditMessage(msg tgbotapi.EditMessageTextConfig) (tgbotapi.Message, error) {
	if msg.Text == "" {
		return tgbotapi.Message{}, errors.New("txt nil")
	}
	mmsg, err := bot.Send(msg)
	if err != nil {
		return mmsg, err
	}
	// go deleteMessage(msg.ChatID, mmsg.MessageID, time.Duration(config.GetInt("redis.cacheTime"))*time.Hour)
	return mmsg, nil
}

func deleteMessage(gid int64, mid int, p time.Duration) {
	time.Sleep(p)
	msgToDelete := tgbotapi.NewDeleteMessage(gid, mid)
	_, _ = bot.Request(msgToDelete)

}
func fillMsgBtn(btnNames []string, btnValues []BtnMessage, msg *tgbotapi.MessageConfig) *tgbotapi.MessageConfig {
	if len(btnNames) != len(btnValues) || len(btnNames) < 1 || msg == nil {
		return msg
	}
	var row []tgbotapi.InlineKeyboardButton
	for k, v := range btnNames {
		btn1 := tgbotapi.NewInlineKeyboardButtonData(v, packBtnMsg(btnValues[k].Type, strings.ToLower((btnValues[k].Value.(string)))))
		row = append(row, btn1)

	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	msg.ReplyMarkup = keyboard
	return msg
}
