package controllers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp-admin/config"
	"github.com/amiraliio/tgbp-admin/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (service *BotService) SetUpChannelByAdmin(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState, text string, userID int) bool {
	if lastState.Data != "" && lastState.State == request.UserState {
		questions := config.QConfig.GetStringMap("SUPERADMIN.CHANNEL.SETUP.QUESTIONS")
		numberOfQuestion := strings.Split(lastState.Data, "_")
		if len(numberOfQuestion) == 2 {
			questioNumber := numberOfQuestion[0]
			relationDate := numberOfQuestion[1]
			prevQuestionNo, err := strconv.Atoi(questioNumber)
			if err == nil {
				tableName := config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N" + questioNumber + ".TABLE_NAME")
				columnName := config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N" + questioNumber + ".COLUMN_NAME")
				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`,`createdAt`) VALUES ('" + tableName + "','" + columnName + "','" + strings.TrimSpace(text) + "','" + strconv.Itoa(userID) + "','" + config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL") + "_" + strconv.Itoa(userID) + "_" + relationDate + "','" + app.CurrentTime + "')")
				if err != nil {
					log.Println(err)
					return true
				}
				if prevQuestionNo+1 > len(questions) {
					service.channelFinalStage(app, bot, relationDate, db, text, userID)
					return true
				}
				service.channelNextQuestion(db, app, bot, m, lastState, relationDate, prevQuestionNo, text, userID)
			}
		}
		return true
	}
	initQuestion := config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N1.QUESTION")
	service.channelSendMessageUserWithActionOnKeyboards(db, app, bot, userID, initQuestion, true)
	SaveUserLastState(db, app, bot, "1_"+strconv.FormatInt(time.Now().Unix(), 10), userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL"))
	return true
}

//next question
func (service *BotService) channelNextQuestion(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, lastState *models.UserLastState, relationDate string, prevQuestionNo int, text string, userID int) {
	questionText := config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".QUESTION")
	if questionText == config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N4.QUESTION") {
		userModel := new(tb.User)
		userModel.ID = userID
		results, err := db.Query("SELECT id,companyName FROM `companies` where type !=''")
		if err != nil {
			log.Println(err)
			return
		}
		defer results.Close()
		replyKeysNestedEven := []tb.ReplyButton{}
		replyKeysNestedOdd := []tb.ReplyButton{}
		var index int
		var hasResult bool
		for results.Next() {
			companymodel := new(models.Company)
			if err := results.Scan(&companymodel.ID, &companymodel.CompanyName); err != nil {
				log.Println(err)
				return
			}
			replyBTN := tb.ReplyButton{
				Text: companymodel.CompanyName,
			}
			if index%2 == 0 {
				replyKeysNestedEven = append(replyKeysNestedEven, replyBTN)
			} else {
				replyKeysNestedOdd = append(replyKeysNestedOdd, replyBTN)
			}
			index++
			hasResult = true
		}
		if !hasResult {
			bot.Send(userModel, config.LangConfig.GetString("MESSAGES.REGISTER_A_COMPANY_FIRST"))
			return
		}
		homeBTN := tb.ReplyButton{
			Text: config.LangConfig.GetString("GENERAL.HOME"),
		}
		replyKeys := [][]tb.ReplyButton{
			replyKeysNestedEven,
			replyKeysNestedOdd,
			[]tb.ReplyButton{homeBTN},
		}
		options := new(tb.SendOptions)
		replyMarkupModel := new(tb.ReplyMarkup)
		replyMarkupModel.ReplyKeyboard = replyKeys
		options.ReplyMarkup = replyMarkupModel
		bot.Send(userModel, questionText, options)
	} else {
		answers := config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".ANSWERS")
		if answers != "" && strings.Contains(strings.TrimSpace(answers), ",") {
			splittedAnswers := strings.Split(answers, ",")
			replyKeysNested := []tb.ReplyButton{}
			for _, v := range splittedAnswers {
				replyBTN := tb.ReplyButton{
					Text: v,
				}
				replyKeysNested = append(replyKeysNested, replyBTN)
			}
			homeBTN := tb.ReplyButton{
				Text: config.LangConfig.GetString("GENERAL.HOME"),
			}
			replyKeys := [][]tb.ReplyButton{
				replyKeysNested,
				[]tb.ReplyButton{homeBTN},
			}
			userModel := new(tb.User)
			userModel.ID = userID
			options := new(tb.SendOptions)
			replyMarkupModel := new(tb.ReplyMarkup)
			replyMarkupModel.ReplyKeyboard = replyKeys
			options.ReplyMarkup = replyMarkupModel
			_, _ = bot.Send(userModel, questionText, options)
		} else {
			userModel := new(tb.User)
			userModel.ID = userID
			homeBTN := tb.ReplyButton{
				Text: config.LangConfig.GetString("GENERAL.HOME"),
			}
			replyKeys := [][]tb.ReplyButton{
				[]tb.ReplyButton{homeBTN},
			}
			options := new(tb.SendOptions)
			replyMarkupModel := new(tb.ReplyMarkup)
			replyMarkupModel.ReplyKeyboard = replyKeys
			options.ReplyMarkup = replyMarkupModel
			bot.Send(userModel, questionText, options)
		}
	}
	SaveUserLastState(db, app, bot, strconv.Itoa(prevQuestionNo+1)+"_"+relationDate, userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL"))
}

func (service *BotService) channelSendMessageUserWithActionOnKeyboards(db *sql.DB, app *config.App, bot *tb.Bot, userID int, message string, showKeyboard bool) {
	userModel := new(tb.User)
	userModel.ID = userID
	homeBTN := tb.ReplyButton{
		Text: config.LangConfig.GetString("GENERAL.HOME"),
	}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{homeBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboardRemove = showKeyboard
	replyModel.ReplyKeyboard = replyKeys
	options := new(tb.SendOptions)
	options.ReplyMarkup = replyModel
	bot.Send(userModel, message, options)
}

func (service *BotService) channelFinalStage(app *config.App, bot *tb.Bot, relationDate string, db *sql.DB, text string, userID int) {
	results, err := db.Query("SELECT id,tableName,columnName,data,relation,status,userID,createdAt from `temp_setup_flow` where status='ACTIVE' and relation=? and userID=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer results.Close()
	if err == nil {
		var uniqueID string
		var channelTableData []*models.TempSetupFlow
		var companyChannelTableData []*models.TempSetupFlow
		var channelsSettings []*models.TempSetupFlow
		for results.Next() {
			tempSetupFlow := new(models.TempSetupFlow)
			err := results.Scan(&tempSetupFlow.ID, &tempSetupFlow.TableName, &tempSetupFlow.ColumnName, &tempSetupFlow.Data, &tempSetupFlow.Relation, &tempSetupFlow.Status, &tempSetupFlow.UserID, &tempSetupFlow.CreatedAt)
			if err != nil {
				log.Println(err)
				return
			}
			switch tempSetupFlow.TableName {
			case config.LangConfig.GetString("GENERAL.COMPANIES_CHANNELS"):
				companyChannelTableData = append(companyChannelTableData, tempSetupFlow)
			case config.LangConfig.GetString("GENERAL.CHANNELS"):
				channelTableData = append(channelTableData, tempSetupFlow)
				if tempSetupFlow.ColumnName == config.QConfig.GetString("SUPERADMIN.CHANNEL.SETUP.QUESTIONS.N2.COLUMN_NAME") {
					uniqueID = tempSetupFlow.Data
				}
			case config.LangConfig.GetString("GENERAL.CHANNELS_SETTINGS"):
				channelsSettings = append(channelsSettings, tempSetupFlow)
			}
		}
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return
		}
		//insert company
		service.insertChannelFinalStateData(app, bot, userID, transaction, channelTableData, companyChannelTableData, channelsSettings, db)
		//update state of temp setup data
		_, err = transaction.Exec("update `temp_setup_flow` set `status`='INACTIVE' where status='ACTIVE' and relation=? and `userID`=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY_CHANNEL")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		channelPublicURL := app.UserBotURL + "?start=join_to_" + uniqueID
		successMessage := config.LangConfig.GetString("MESSAGES.CHANNEL_REGISTERED_SUCCESSFULLY_WITH_URL") + channelPublicURL
		_, err = transaction.Exec("update `channels` set `publicURL`=? where uniqueID=?", channelPublicURL, uniqueID)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		err = transaction.Commit()
		if err != nil {
			log.Println(err)
			return
		}
		service.channelSendMessageUserWithActionOnKeyboards(db, app, bot, userID, successMessage, false)
		SaveUserLastState(db, app, bot, text, userID, config.LangConfig.GetString("STATE.DONE_SETUP_VERIFIED_COMPANY_CHANNEL"))
	}
}

func (service *BotService) insertChannelFinalStateData(app *config.App, bot *tb.Bot, userID int, transaction *sql.Tx, channelTableData, companyChannelTableData, channelsSettings []*models.TempSetupFlow, db *sql.DB) {
	if companyChannelTableData == nil || channelTableData == nil || channelsSettings == nil {
		transaction.Rollback()
		log.Println(config.LangConfig.GetString("MESSAGES.DATA_MUST_NOT_BE_NULL"))
		return
	}

	//select company id
	var companyName string
	for _, v := range companyChannelTableData {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.COMPANY_ID") {
			companyName = v.Data
		}
	}
	companyNewModel := new(models.Company)
	var companyID int64
	if err := db.QueryRow("SELECT id,companyName FROM `companies` where `companyName`=?", companyName).Scan(&companyNewModel.ID, &companyNewModel.CompanyName); err != nil {
		_ = transaction.Rollback()
		userModel := new(tb.User)
		userModel.ID = userID
		bot.Send(userModel, config.LangConfig.GetString("MESSAGES.COMPANY_NOT_EXIST"))
		return
	} else {
		companyID = companyNewModel.ID
	}

	//insert channel
	var manualChannelName, uniqueID, channelURL string
	for _, v := range channelTableData {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.MANUAL_CHANNEL_NAME") {
			manualChannelName = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.UNIQUE_ID") {
			uniqueID = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.CHANNEL_URL") {
			channelURL = v.Data
		}
	}
	channelModel := new(models.Channel)
	if err := db.QueryRow("SELECT channelID,id FROM `channels` where `uniqueID`=?", uniqueID).Scan(&channelModel.ChannelID, &channelModel.ID); err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	_, err := transaction.Exec("update `channels` set `manualChannelName`=?,  `channelURL`=? where `uniqueID`=?", manualChannelName, channelURL, uniqueID)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//remove previous companies_channels, which create with channel id
	_, err = transaction.Exec("delete from `companies_channels` where `channelID`='" + strconv.FormatInt(channelModel.ID, 10) + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert company channel
	_, err = transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + strconv.FormatInt(companyID, 10) + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + app.CurrentTime + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//remove previous company, which create with channel id
	_, err = transaction.Exec("delete from `companies` where `companyName`='" + channelModel.ChannelID + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert channel settings
	var joinVerify, newMessageVerify, replyVerify, directVerify string
	for _, v := range channelsSettings {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.JOIN_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				joinVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				joinVerify = "0"
			}
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.NEW_MESSAGE_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				newMessageVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				newMessageVerify = "0"
			}
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.REPLY_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				replyVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				replyVerify = "0"
			}
		}
		if v.ColumnName == "directVerify" {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				directVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				directVerify = "0"
			}
		}
	}

	//TODO add transactional
	//remove previous company, which create with channel id
	_, _ = transaction.Exec("delete from `channels_settings` where `channelID`=?", channelModel.ID)

	_, err = transaction.Exec("INSERT INTO `channels_settings` (`joinVerify`,`newMessageVerify`,`replyVerify`,`directVerify`,`channelID`,`createdAt`) VALUES(?,?,?,?,?,?)", joinVerify, newMessageVerify, replyVerify, directVerify, strconv.FormatInt(channelModel.ID, 10), app.CurrentTime)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
}
