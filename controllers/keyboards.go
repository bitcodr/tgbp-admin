//Package controllers ...
package controllers

import (
	tb "gopkg.in/tucnak/telebot.v2"
)


//bot startup buttons
var setupVerifiedCompany = tb.ReplyButton{
	Text: "Setup Verified Company Account",
}
var setupVerifiedCompanyChannel = tb.ReplyButton{
	Text: "Setup Verified Company Channel",
}

var StartBotKeys = [][]tb.ReplyButton{
	[]tb.ReplyButton{setupVerifiedCompany},
	[]tb.ReplyButton{setupVerifiedCompanyChannel},
}
