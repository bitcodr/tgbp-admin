package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp-admin/config"
	"github.com/amiraliio/tgbp-admin/helpers"
	"github.com/amiraliio/tgbp-admin/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onCallbackEvents(app *config.App, bot *tb.Bot) {
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {

		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)

		//check incoming text
		incomingMessage := c.Data
		switch {
		case incomingMessage == config.LangConfig.GetString("GENERAL.HOME") || incomingMessage == config.LangConfig.GetString("COMMANDS.START"):
			goto StartBotCallback
		default:
			goto CheckState
		}

	StartBotCallback:
		if onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  config.LangConfig.GetString("STATE.HOME"),
			Command:    config.LangConfig.GetString("GENERAL.HOME"),
			Command1:   config.LangConfig.GetString("COMMANDS.START"),
			Controller: "StartBotCallback",
		}) {
			Init(app, bot, true)
		}
		goto END

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		switch lastState.State {
		case config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"):
			goto SetUpCompanyByAdmin
		case config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL"):
			goto SetUpChannelByAdmin
		default:
			bot.Send(c.Sender, "Your message "+c.Data+" is not being processed or sent to any individual, channel or group. Please use inline buttons or use the /home command.")
			goto END
		}

	SetUpCompanyByAdmin:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"),
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	SetUpChannelByAdmin:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL"),
			Controller: "SetUpChannelByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	END:
	})
}

func onCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, request *Event) bool {
	var result bool
	helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	return result
}

func inlineOnCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, db *sql.DB, lastState *models.UserLastState, request *Event) bool {
	var result bool
	switch {
	case request.Controller == "SetUpChannelByAdmin" || request.Controller == "SetUpCompanyByAdmin":
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
	default:
		helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	}
	return result
}
