package main

import (
	"fmt"
	"regexp"

	// "log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processUserCommand(update *tgbotapi.Update) {
	// var msg tgbotapi.MessageConfig
	upmsg := update.Message
	// msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")

	switch upmsg.Command() {
	case "start", "help":
		s_start(*update, false)
	case "checkin":
		s_checkin(update.Message.Chat.ID, update.Message.From.ID)
	case "bind":
		s_bind(*update)
	}
}
func processUserBtnCallBack(update *tgbotapi.Update) {
	if update.CallbackQuery != nil {
		// 获取按钮的回调数据
		data, err := unPackBtnMsg(update.CallbackQuery.Data)
		if err != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error())
			bot.Send(msg)
		}
		// 根据不同的回调数据进行不同的操作
		switch data.Type {

		case BIND:
			pre_bind(*update)
		case UNBIND:
			s_unbind(*update)
		case INVITE:
			s_invite(*update)
		case ACCOUNT:
			s_account(*update)
		case SHOP:
			s_shop(*update)
		case CHECKIN:
			s_checkin(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)
		case BACK:
			s_start(*update, true)
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
func processUserTxt(update *tgbotapi.Update) {
	if update.Message.Text == "323232" {

	} else if update.Message.Text == "434324234" {

	} else {
		forwardMsgToAdmins(update)
	}
}

func forwardMsgToAdmins(update *tgbotapi.Update) {
	for _, id := range config.Get("telegram.admins").([]interface{}) {
		fm := tgbotapi.NewForward(int64(id.(int)), update.Message.Chat.ID, update.Message.MessageID)
		m, e := bot.Send(fm)
		if e == nil {
			redisClient.Set(strconv.Itoa(m.MessageID)+"forward", update.Message.Chat.ID, 0)
		} else {
			info := fmt.Sprintf("tg://user?id=%d", update.Message.From.ID)
			ffm := tgbotapi.NewCopyMessage(int64(id.(int)), update.Message.Chat.ID, update.Message.MessageID)
			ffm.Caption = fmt.Sprintf("\n来自: [%s](%s)\n", update.Message.From.UserName, info)
			mm, ee := bot.Send(ffm)
			if ee == nil {
				redisClient.Set(strconv.Itoa(mm.MessageID)+"forward", update.Message.Chat.ID, 0)
			}
		}
	}
}

func s_checkin(chatId int64, userId int64, msgId ...int) {
	user := QueryUser(userId)
	if len(msgId) > 0 {
		mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, msgId[0], "", tgbotapi.InlineKeyboardMarkup{})
		if user.Id <= 0 {
			mm.Text = "⛔️当前未绑定账户\n请发送 /bind <订阅地址> 绑定账户\n\n#示例\n/bind https://域名/api/v1/client/subscribe?token=c09a65fd29cb8453926642c0db2e74c0"
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendEditMessage(mm)
			return
		}
		if user.PlanId <= 0 {
			mm.Text = "⛔当前暂无订阅计划,请购买后才能签到赚取流量😯..."
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendEditMessage(mm)
			return
		}

		cc := CheckinTime(userId)
		if cc == false {
			mm.Text = fmt.Sprintf("🥳今天已经签到过啦...")
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendEditMessage(mm)
			return
		}

		uu := checkinUser(userId)

		mm.Text = fmt.Sprintf("💍签到成功\n本次签到获得 %s 流量\n下次签到时间: %s", ByteSize(uu.CheckinTraffic), UnixToStr(uu.NextAt))
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
	} else {
		mm := tgbotapi.NewMessage(chatId, "")
		if user.Id <= 0 {
			mm.Text = "⛔️当前未绑定账户\n请发送 /bind <订阅地址> 绑定账户\n\n#示例\n/bind https://域名/api/v1/client/subscribe?token=c09a65fd29cb8453926642c0db2e74c0"
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendMessage(mm)
			return
		}
		if user.PlanId <= 0 {
			mm.Text = "⛔当前暂无订阅计划,请购买后才能签到赚取流量😯..."
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendMessage(mm)
			return
		}

		cc := CheckinTime(userId)
		if cc == false {
			mm.Text = fmt.Sprintf("🥳今天已经签到过啦...")
			mm.ReplyMarkup = getBackKeyboard()
			_, _ = sendMessage(mm)
			return
		}

		uu := checkinUser(userId)

		mm.Text = fmt.Sprintf("💍签到成功\n本次签到获得 %s 流量\n下次签到时间: %s", ByteSize(uu.CheckinTraffic), UnixToStr(uu.NextAt))
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendMessage(mm)
	}

}

func s_account(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	userId := update.CallbackQuery.From.ID
	user := QueryUser(userId)
	mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, update.CallbackQuery.Message.MessageID, "", tgbotapi.InlineKeyboardMarkup{})
	if user.Id <= 0 {
		mm.Text = "⛔️当前未绑定账户\n请发送 /bind <订阅地址> 绑定账户\n\n#示例\n/bind https://域名/api/v1/client/subscribe?token=c09a65fd29cb8453926642c0db2e74c0"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	p := QueryPlan(int(user.PlanId))
	Email := user.Email
	CreatedAt := UnixToStr(user.CreatedAt)
	Balance := user.Balance / 100
	CommissionBalance := user.CommissionBalance / 100
	PlanName := p.Name
	ExpiredAt := UnixToStr(user.ExpiredAt)
	TransferEnable := ByteSize(user.TransferEnable)
	U := ByteSize(user.U)
	D := ByteSize(user.D)
	S := ByteSize(user.TransferEnable - (user.U + user.D))
	if user.PlanId <= 0 {
		mm.Text = fmt.Sprintf("🧚🏻账户信息概况:\n\n当前绑定账户: %s\n注册时间: %s\n账户余额: %d元\n佣金余额: %d元\n\n当前订阅: 当前暂无订阅计划", Email, CreatedAt, Balance, CommissionBalance)
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}

	mm.Text = fmt.Sprintf("🧚🏻账户信息概况:\n\n当前绑定账户: %s\n注册时间: %s\n账户余额: %d元\n佣金余额: %d元\n\n当前订阅: %s\n到期时间: %s\n订阅流量: %s\n已用上行: %s\n已用下行: %s\n剩余可用: %s", Email, CreatedAt, Balance, CommissionBalance, PlanName, ExpiredAt, TransferEnable, U, D, S)
	mm.ReplyMarkup = getBackKeyboard()
	_, _ = sendEditMessage(mm)

}

func s_bind(update tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	userId := update.Message.From.ID
	user := QueryUser(userId)
	mm := tgbotapi.NewMessage(chatId, "")
	if user.Id > 0 {
		mm.Text = fmt.Sprintf("⭐您当前绑定账户: %s\n若需要修改绑定,请先解绑当前账户！", user.Email)
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendMessage(mm)
		return
	}

	format := strings.Index(update.Message.Text, "token=")
	if format <= 0 {
		mm.Text = "⭐️️账户绑定格式: /bind <订阅地址>\n\n 发送示例：\n/bind https://域名/api/v1/client/subscribe?token=c09a65fd29cb8453926642c0db2e74c0"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendMessage(mm)
		return
	}

	b := BindUser(update.Message.Text[format:], update.Message.Chat.ID)
	if b.Id <= 0 {
		mm.Text = "❌订阅无效,请前往官网复制最新订阅地址!"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendMessage(mm)
		return
	}

	if b.TelegramId != uint(update.Message.Chat.ID) {
		mm.Text = "❌账户绑定失败,请稍后再试"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendMessage(mm)
		return
	}
	mm.Text = fmt.Sprintf("💍账户绑定成功: %s", b.Email)
	mm.ReplyMarkup = getBackKeyboard()
	_, _ = sendMessage(mm)
}

func s_unbind(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	userId := update.CallbackQuery.From.ID
	user := unbindUser(userId)
	mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, update.CallbackQuery.Message.MessageID, "", tgbotapi.InlineKeyboardMarkup{})
	if user.Id <= 0 {
		mm.Text = "⛔️当前未绑定账户"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	if user.TelegramId > 0 {
		mm.Text = "❌账户解绑失败,请稍后再试..."
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	mm.Text = "🪖账户解绑成功"
	mm.ReplyMarkup = getBackKeyboard()
	_, _ = sendEditMessage(mm)
}
func s_invite(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	userId := update.CallbackQuery.From.ID
	user := QueryUser(userId)
	mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, update.CallbackQuery.Message.MessageID, "", tgbotapi.InlineKeyboardMarkup{})
	if user.Id <= 0 {
		mm.Text = "❌订阅无效,请前往官网复制最新订阅地址!"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
	}
	invites := getInviteList(userId)
	if len(invites) <= 0 {
		mm.Text = "⛔️当前暂无邀请记录"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	msg := ""
	total := 0
	directNum := 0
	if len(invites) > 1 {
		msg += "其中我邀请的\n"
	}
	j :=3
	for k, v := range invites {
		if k < j {
			if  v.InviteUserId == user.Id{
				j+=1
			}else{
				u := getUserById(v.InviteUserId)
				msg += fmt.Sprintf("👉🏻`%s`邀请了%v人\n", u.Email, v.Num)
			}
		}
		if v.InviteUserId == user.Id {
			directNum = v.Num
		}
		total += v.Num
	}

	fm := fmt.Sprintf("🧚🏻邀请信息:\n\n生态影响: %d人\n直接邀请: %d人\n间接邀请: %d人\n%v\n我的邀请链接:\n `%s`", total, directNum, total-directNum, msg, getInviteLink(user.Id))
	mm.Text = fm
	mm.DisableWebPagePreview = false
	btn1 := tgbotapi.NewInlineKeyboardButtonData("↩️返回主菜单", packBtnMsg(BACK, BACK))
	row := tgbotapi.NewInlineKeyboardRow(btn1)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	mm.ReplyMarkup = &keyboard
	mm.ParseMode = tgbotapi.ModeMarkdown
	_, _ = sendEditMessage(mm)
}

func s_shop(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	userId := update.CallbackQuery.From.ID
	plans := getPlanList()
	mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, update.CallbackQuery.Message.MessageID, "", tgbotapi.InlineKeyboardMarkup{})
	if len(plans) <= 0 {
		mm.Text = "⛔️当前暂无订阅计划"
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	msg := "🔖<b>套餐列表</b>\n\n"
	for _, v := range plans {
		re := regexp.MustCompile(`(?i)<p[^>]*>`)

		v.Content = strings.ReplaceAll(v.Content, "<br>", "\n")
		v.Content = strings.ReplaceAll(v.Content, "<br/>", "\n")
		v.Content = re.ReplaceAllString(v.Content, "")
		v.Content = strings.ReplaceAll(v.Content, "</p>", " ")
		msg += "🌲🌲🌲🌲🌲🌲🌲🌲🌲🌲🌲🌲🌲\n"
		if v.MonthPrice > 0 {
			v.MonthPrice = v.MonthPrice / 100
			msg += fmt.Sprintf("🎁<b>%v\n\n月付: ￥%v</b>\n\n%v\n\n", v.Name, v.MonthPrice, v.Content)
		} else if v.OnetimePrice > 0 {
			v.OnetimePrice = v.OnetimePrice / 100
			msg += fmt.Sprintf("🎁<b>%v\n\n一次性: ￥%v</b>\n\n%v\n\n", v.Name, v.OnetimePrice, v.Content)
		}
		mm.Text = msg
	}
	btn1 := tgbotapi.NewInlineKeyboardButtonData("↩️返回主菜单", packBtnMsg(BACK, BACK))
	url := getTelegramLoginUrl(userId, "plan")
	var row1 []tgbotapi.InlineKeyboardButton
	if url != "" {
		btn2 := tgbotapi.NewInlineKeyboardButtonURL("💰前往购买", getTelegramLoginUrl(userId, "plan"))
		row1 = tgbotapi.NewInlineKeyboardRow(btn1, btn2)
	} else {
		row1 = tgbotapi.NewInlineKeyboardRow(btn1)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1)
	mm.ReplyMarkup = &keyboard
	mm.ParseMode = tgbotapi.ModeHTML
	_, _ = sendEditMessage(mm)
}

func getTelegramLoginUrl(id int64, redirect string) string {
	user := QueryUser(id)
	if user.Id <= 0 {
		return ""
	}
	url := config.GetString("websiteUrl") + "/#/login?verify=" + user.Token + "&redirect=" + redirect
	return url
}

func s_start(update tgbotapi.Update, isBack bool) {

	btn1 := tgbotapi.NewInlineKeyboardButtonData("💍绑定账户", packBtnMsg(BIND, BIND))
	btn2 := tgbotapi.NewInlineKeyboardButtonData("🪖解绑账户", packBtnMsg(UNBIND, UNBIND))
	btn3 := tgbotapi.NewInlineKeyboardButtonData("🎉签到", packBtnMsg(CHECKIN, CHECKIN))
	btn4 := tgbotapi.NewInlineKeyboardButtonData("🧑‍🍼个人信息", packBtnMsg(ACCOUNT, ACCOUNT))
	btn5 := tgbotapi.NewInlineKeyboardButtonData("💁我的邀请", packBtnMsg(INVITE, INVITE))
	btn6 := tgbotapi.NewInlineKeyboardButtonData("🛒商店", packBtnMsg(SHOP, SHOP))
	btn7 := tgbotapi.NewInlineKeyboardButtonURL("🌐前往官网", config.GetString("websiteUrl"))
	btn8 := tgbotapi.NewInlineKeyboardButtonURL("💞加入TG群", config.GetString("tgGroupLink"))
	row1 := tgbotapi.NewInlineKeyboardRow(btn1, btn2)
	row2 := tgbotapi.NewInlineKeyboardRow(btn3, btn4,btn5)
	row3 := tgbotapi.NewInlineKeyboardRow(btn6, btn7,btn8)
	grow := tgbotapi.NewInlineKeyboardRow(btn1, btn2)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3)
	gk := tgbotapi.NewInlineKeyboardMarkup(grow)
	prvTxt := fmt.Sprintf("🤖V2board机器人\n\n欢迎使用%v,您可通过向此bot发送消息,客服将会收到您的反馈并通过此回复。", config.GetString("appName"))
	grpTxt := fmt.Sprintf("🤖V2board机器人\n\n欢迎使用%v", config.GetString("appName"))
	if isBack {
		msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "", keyboard)
		if update.FromChat().IsPrivate() {
			msg.Text = prvTxt
		} else {
			msg.Text = grpTxt
			if config.Get("isAutoDeleteMsg").(bool) {
				msgToDelete := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				_, _ = bot.Request(msgToDelete)
			}
		}
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.FromChat().IsPrivate() {
			msg.Text = prvTxt
			msg.ReplyMarkup = keyboard
		} else {
			msg.Text = grpTxt
			msg.ReplyMarkup = gk
			if config.Get("isAutoDeleteMsg").(bool) {
				msgToDelete := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
				_, _ = bot.Request(msgToDelete)
			}
		}
		
		msg.ParseMode = "Markdown"
		sendMessage(msg)
	}
}

func getBackKeyboard() *tgbotapi.InlineKeyboardMarkup {
	btn1 := tgbotapi.NewInlineKeyboardButtonData("↩️返回主菜单", packBtnMsg(BACK, BACK))
	row1 := tgbotapi.NewInlineKeyboardRow(btn1)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1)
	return &keyboard
}

func pre_bind(update tgbotapi.Update){
	chatId := update.CallbackQuery.Message.Chat.ID
	userId := update.CallbackQuery.From.ID
	user := QueryUser(userId)
	mm := tgbotapi.NewEditMessageTextAndMarkup(chatId, update.CallbackQuery.Message.MessageID, "", tgbotapi.InlineKeyboardMarkup{})
	if user.Id > 0 {
		mm.Text = fmt.Sprintf("⭐您当前已经绑定账户: %s\n若需要修改绑定,请先解绑当前账户！", user.Email)
		mm.ReplyMarkup = getBackKeyboard()
		_, _ = sendEditMessage(mm)
		return
	}
	mm.Text ="⭐️️账户绑定格式: /bind <订阅地址>\n\n 发送示例：\n/bind https://域名/api/v1/client/subscribe?token=c09a65fd29cb8453926642c0db2e74c0"
	mm.ReplyMarkup =getBackKeyboard()
	sendEditMessage(mm)
}
