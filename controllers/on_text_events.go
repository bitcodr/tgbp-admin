package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp-admin/config"
	"github.com/amiraliio/tgbp-admin/helpers"
	"github.com/amiraliio/tgbp-admin/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onTextEvents(app *config.App, bot *tb.Bot) {

	bot.Handle(tb.OnText, func(message *tb.Message) {
		if !message.Private() {
			return
		}

		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)

		//check incoming text
		incomingMessage := message.Text
		switch {
		case incomingMessage == config.LangConfig.GetString("GENERAL.HOME") || incomingMessage == config.LangConfig.GetString("COMMANDS.START"):
			goto StartBot
		default:
			goto CheckState
		}

	StartBot:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  config.LangConfig.GetString("STATE.HOME"),
			Command:    config.LangConfig.GetString("GENERAL.HOME"),
			Command1:   config.LangConfig.GetString("COMMANDS.START"),
			Controller: "StartBot",
		}) {
			Init(app, bot, true)
		}
		goto END

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		switch {
		case lastState.State == config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY") || incomingMessage == setupVerifiedCompany.Text:
			goto SetUpCompanyByAdmin
		case lastState.State == config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL") || incomingMessage == setupVerifiedCompanyChannel.Text:
			goto SetUpChannelByAdmin
		default:
			bot.Send(message.Sender, "Your message "+message.Text+" is not being processed or sent to any individual, channel or group. Please use inline buttons or use the /home command.")
			goto END
		}

	SetUpCompanyByAdmin:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"),
			Command:    setupVerifiedCompany.Text,
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	SetUpChannelByAdmin:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL"),
			Command:    setupVerifiedCompany.Text,
			Controller: "SetUpChannelByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	END:
	})
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, db *sql.DB, lastState *models.UserLastState, request *Event) bool {
	var result bool
	switch {
	case request.Controller == "SetUpCompanyByAdmin" || request.Controller == "SetUpChannelByAdmin":
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
	default:
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState)
	}
	return result
}
