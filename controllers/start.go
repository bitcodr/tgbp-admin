package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp-admin/config"
	"github.com/amiraliio/tgbp-admin/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
)

func (service *BotService) StartBot(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) bool {
	if strings.TrimSpace(message.Text) == request.Command || strings.TrimSpace(message.Text) == request.Command1 {
		db := app.DB()
		defer db.Close()
		if message.Sender != nil {
			SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.UserState)
		}
		newReplyModel := new(tb.ReplyMarkup)
		newReplyModel.ReplyKeyboard = StartBotKeys
		newSendOption := new(tb.SendOptions)
		newSendOption.ReplyMarkup = newReplyModel
		_ = bot.Delete(message)
		_, err := bot.Send(message.Sender, config.LangConfig.GetString("MESSAGES.WHAT_TO_YOU_WANT"), newSendOption)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func (service *BotService) StartBotCallback(app *config.App, bot *tb.Bot, callback *tb.Callback, request *Event) bool {
	if strings.TrimSpace(callback.Data) == request.Command || strings.TrimSpace(callback.Data) == request.Command1 {
		db := app.DB()
		defer db.Close()
		if callback.Sender != nil {
			SaveUserLastState(db, app, bot, callback.Data, callback.Sender.ID, request.UserState)
		}
		newReplyModel := new(tb.ReplyMarkup)
		newReplyModel.ReplyKeyboard = StartBotKeys
		newSendOption := new(tb.SendOptions)
		newSendOption.ReplyMarkup = newReplyModel
		_ = bot.Delete(callback.Message)
		_, err := bot.Send(callback.Sender, config.LangConfig.GetString("MESSAGES.WHAT_TO_YOU_WANT"), newSendOption)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func SaveUserLastState(db *sql.DB, app *config.App, bot *tb.Bot, data string, userDataID int, state string) {
	userID := strconv.Itoa(userDataID)
	insertedState, err := db.Query("INSERT INTO `users_last_state` (`userID`,`state`,`data`,`createdAt`) VALUES('" + userID + "','" + state + "','" + strings.TrimSpace(data) + "','" + app.CurrentTime + "')")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertedState.Close()
}

func GetUserLastState(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, user int) *models.UserLastState {
	userLastState := new(models.UserLastState)
	if err := db.QueryRow("SELECT `data`,`state`,`userID` from `users_last_state` where `userId`=? order by `id` DESC limit 1", user).Scan(&userLastState.Data, &userLastState.State, &userLastState.UserID); err != nil {
		log.Println(err)
		userLastState.Status = "INACTIVE"
		return userLastState
	}
	return userLastState
}
