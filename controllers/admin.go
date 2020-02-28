package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp-admin/helpers"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp-admin/config"
	"github.com/amiraliio/tgbp-admin/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

//TODO add gmail.com and yahoo

func (service *BotService) SetUpCompanyByAdmin(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState, text string, userID int) bool {
	if lastState.Data != "" && lastState.State == request.UserState {

		questions := config.QConfig.GetStringMap("SUPERADMIN.COMPANY.SETUP.QUESTIONS")
		numberOfQuestion := strings.Split(lastState.Data, "_")

		if len(numberOfQuestion) == 2 {

			questioNumber := numberOfQuestion[0]
			relationDate := numberOfQuestion[1]
			prevQuestionNo, err := strconv.Atoi(questioNumber)

			if err == nil {

				botUserModel := new(tb.User)
				botUserModel.ID = userID

				tableName := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".TABLE_NAME")
				columnName := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".COLUMN_NAME")

				if columnName == config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.COLUMN_NAME") {
					if strings.Contains(text, " ") {
						bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.COMPANY_NAME_CANNOT_HAVE_SPACES"))
						return true
					}
					companyModel := new(models.Company)
					if err := db.QueryRow("SELECT id FROM `companies` where companyName=?", text).Scan(&companyModel.ID); err == nil {
						bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.COMPANY_WITH_THE_NAME_EXIST_CHOOSE_ANOTHER_ONE"))
						return true
					}
				}

				if columnName == config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N2.COLUMN_NAME") {

					if strings.Contains(strings.TrimSpace(text), ",") {
						suffixes := strings.Split(strings.TrimSpace(text), ",")
						for _, suffix := range suffixes {
							if !strings.Contains(suffix, "@") {
								bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.PLEASE_ENTER_VALID_EMAIL_SUFFIX"))
								return true
							}
							emails := []string{"@hotmail.com", "@outlook.com", "@zoho.com", "@icloud.com", "@mail.com", "@aol.com", "@yandex.com"}
							if helpers.SortAndSearchInStrings(emails, suffix) {
								bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.NOT_ALLOWED_PUBLIC_EMAIL_SUFFIX")+suffix)
								return true
							}
							suffixesModel := new(models.CompanyEmailSuffixes)
							err := db.QueryRow("SELECT id FROM `companies_email_suffixes` where `suffix`=?", suffix).Scan(&suffixesModel.Suffix)
							if err == nil {
								bot.Send(botUserModel, suffix+config.LangConfig.GetString("MESSAGES.EMAIL_SUFFIX_EXIST"))
								return true
							}
						}

					} else {

						botUserModel := new(tb.User)
						botUserModel.ID = userID
						if !strings.Contains(text, "@") {
							bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.PLEASE_ENTER_VALID_EMAIL_SUFFIX"))
							return true
						}
						emails := []string{"@hotmail.com", "@outlook.com", "@zoho.com", "@icloud.com", "@mail.com", "@aol.com", "@yandex.com"}
						if helpers.SortAndSearchInStrings(emails, text) {
							bot.Send(botUserModel, config.LangConfig.GetString("MESSAGES.NOT_ALLOWED_PUBLIC_EMAIL_SUFFIX")+text)
							return true
						}
						suffixesModel := new(models.CompanyEmailSuffixes)
						err := db.QueryRow("SELECT id FROM `companies_email_suffixes` where `suffix`=?", strings.TrimSpace(text)).Scan(&suffixesModel.Suffix)
						if err == nil {
							bot.Send(botUserModel, strings.TrimSpace(text)+config.LangConfig.GetString("MESSAGES.EMAIL_SUFFIX_EXIST"))
							return true
						}
					}
				}

				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`,`createdAt`) VALUES ('" + tableName + "','" + columnName + "','" + strings.TrimSpace(text) + "','" + strconv.Itoa(userID) + "','" + config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY") + "_" + strconv.Itoa(userID) + "_" + relationDate + "','" + app.CurrentTime + "')")
				if err != nil {
					log.Println(err)
					return true
				}

				if prevQuestionNo+1 > len(questions) {
					service.finalStage(app, bot, relationDate, db, text, userID)
					return true
				}

				service.nextQuestion(db, app, bot, m, lastState, relationDate, prevQuestionNo, text, userID)
			}
		}
		return true
	}
	initQuestion := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.QUESTION")
	service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, initQuestion, true)
	SaveUserLastState(db, app, bot, "1_"+strconv.FormatInt(time.Now().Unix(), 10), userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"))
	return true
}

//next question
func (service *BotService) nextQuestion(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, lastState *models.UserLastState, relationDate string, prevQuestionNo int, text string, userID int) {

	questionText := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".QUESTION")
	answers := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".ANSWERS")

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

	SaveUserLastState(db, app, bot, strconv.Itoa(prevQuestionNo+1)+"_"+relationDate, userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"))
}

func (service *BotService) sendMessageUserWithActionOnKeyboards(db *sql.DB, app *config.App, bot *tb.Bot, userID int, message string, showKeyboard bool) {
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

func (service *BotService) finalStage(app *config.App, bot *tb.Bot, relationDate string, db *sql.DB, text string, userID int) {
	results, err := db.Query("SELECT id,tableName,columnName,data,relation,status,userID,createdAt from `temp_setup_flow` where status='ACTIVE' and relation=? and userID=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer results.Close()
	if err == nil {
		var companyTableData []*models.TempSetupFlow
		var companiesEmailSuffixes []*models.TempSetupFlow
		var companyName string
		for results.Next() {
			tempSetupFlow := new(models.TempSetupFlow)
			err := results.Scan(&tempSetupFlow.ID, &tempSetupFlow.TableName, &tempSetupFlow.ColumnName, &tempSetupFlow.Data, &tempSetupFlow.Relation, &tempSetupFlow.Status, &tempSetupFlow.UserID, &tempSetupFlow.CreatedAt)
			if err != nil {
				log.Println(err)
				return
			}
			switch tempSetupFlow.TableName {
			case config.LangConfig.GetString("GENERAL.COMPANIES"):
				companyTableData = append(companyTableData, tempSetupFlow)
				if tempSetupFlow.ColumnName == config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.COLUMN_NAME") {
					companyName = tempSetupFlow.Data
				}
			case config.LangConfig.GetString("GENERAL.COMPANY_EMAIL_SUFFIXES"):
				companiesEmailSuffixes = append(companiesEmailSuffixes, tempSetupFlow)
			}
		}
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return
		}
		//insert company
		service.insertFinalStateData(app, bot, userID, transaction, companyTableData, companiesEmailSuffixes, db)
		//update state of temp setup data
		_, err = transaction.Exec("update `temp_setup_flow` set `status`='INACTIVE' where status='ACTIVE' and relation=? and `userID`=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		companyPublicURL := app.UserBotURL + "?start=join_company_" + companyName
		successMessage := config.LangConfig.GetString("MESSAGES.COMPANY_REGISTERED_SUCCESSFULLY") + companyPublicURL

		_, err = transaction.Exec("update `companies` set `publicURL`=? where companyName=?", companyPublicURL, companyName)
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

		service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, successMessage, false)
		SaveUserLastState(db, app, bot, text, userID, config.LangConfig.GetString("STATE.DONE_SETUP_VERIFIED_COMPANY"))
	}
}

func (service *BotService) insertFinalStateData(app *config.App, bot *tb.Bot, userID int, transaction *sql.Tx, companyTableData, companiesEmailSuffixes []*models.TempSetupFlow, db *sql.DB) {
	if companyTableData == nil || companiesEmailSuffixes == nil || len(companiesEmailSuffixes) != 1 {
		transaction.Rollback()
		log.Println(config.LangConfig.GetString("MESSAGES.DATA_MUST_NOT_BE_NULL"))
		return
	}
	//insert company
	var companyName string
	for _, v := range companyTableData {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.COMPANY_NAME") {
			companyName = v.Data
		}
	}
	companyNewModel := new(models.Company)
	var companyID int64
	if err := db.QueryRow("SELECT id,companyName FROM `companies` where `companyName`=?", companyName).Scan(&companyNewModel.ID, &companyNewModel.CompanyName); err != nil {
		insertCompany, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`createdAt`) VALUES(?,?)", companyName, app.CurrentTime)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		companyID, err = insertCompany.LastInsertId()
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
	} else {
		companyID = companyNewModel.ID
	}

	//insert channelsEmailSuffixes
	emailSuffixed := companiesEmailSuffixes[0]
	if strings.Contains(emailSuffixed.Data, ",") {
		suffixes := strings.Split(emailSuffixed.Data, ",")
		for _, suffix := range suffixes {
			_, err := transaction.Exec("INSERT INTO `companies_email_suffixes` (`suffix`,`companyID`,`createdAt`) VALUES(?,?,?)", suffix, strconv.FormatInt(companyID, 10), app.CurrentTime)
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return
			}
		}
	} else {
		_, err := transaction.Exec("INSERT INTO `companies_email_suffixes` (`suffix`,`companyID`,`createdAt`) VALUES(?,?,?)", emailSuffixed.Data, strconv.FormatInt(companyID, 10), app.CurrentTime)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
	}
}
