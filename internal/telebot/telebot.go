package telebot

import (
	"fmt"
	"sync"

	"github.com/bsergik/telebot-nats/internal/database"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	tele "gopkg.in/tucnak/telebot.v3"
)

/******************************************************************************
 * Types
 */

type Telebot struct {
	Zaplog        *zap.Logger
	TelegramToken string
	MasterUserID  int
	Storage       database.IStorage
}

/******************************************************************************
 * Global variables
 */

var (
	_gZaplogSugar     *zap.SugaredLogger
	_gTelegramToken   string
	_gMasterUserID    int
	_gTelebot         *tele.Bot
	_gTelebotUsername string
	_gRecipients      recipients
	_gStorage         database.IStorage
)

/******************************************************************************
 * Methods
 */

func (t *Telebot) Start(wg *sync.WaitGroup, finish <-chan struct{}) {
	t.init()

	go _gTelebot.Start()

	_gZaplogSugar.Info("Started.")

	<-finish

	_gZaplogSugar.Info("Gracefully stopped.")
	_gTelebot.Stop()
	wg.Done()
}

func (t *Telebot) init() {
	_gZaplogSugar = t.Zaplog.Sugar()
	_gTelegramToken = t.TelegramToken
	_gMasterUserID = t.MasterUserID
	_gStorage = t.Storage

	_gRecipients.reset()
	_gTelebot = t.initTelebot()

	_gRecipients.add(&tele.User{ID: t.MasterUserID})
	fillInRecipientsFromDB()
}

func (t *Telebot) initTelebot() *tele.Bot {
	telebot, err := tele.NewBot(tele.Settings{
		Token: t.TelegramToken,
		Poller: &tele.LongPoller{
			Limit:        0,
			Timeout:      tbPolletTimeout,
			LastUpdateID: 0,
			AllowedUpdates: []string{
				"message",
				"inline_query",
				"callback_query",
				"chosen_inline_result",
			},
		},
	})

	if err != nil {
		_gZaplogSugar.Fatal(err)
	}

	_gTelebotUsername = telebot.Me.Username

	telebot.Handle("/inviteme", invitemeHandler)
	telebot.Handle("/forgetme", forgetmeHandler)
	telebot.Handle(tele.OnCallback, onCallbackHandler)

	return telebot
}

func (t *Telebot) SendAchtungMessage(msg *database.Message) (resErr error) {
	_gRecipients.mutex.Lock()

	defer _gRecipients.mutex.Unlock()

	for i := range _gRecipients.rcpns {
		mOut, err := _gTelebot.Send(
			_gRecipients.rcpns[i],
			formatMessage(msg),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown},
		)
		if err != nil {
			_gZaplogSugar.Errorf("Cannot send message. Error: %s", err)

			resErr = fmt.Errorf("%s. Error: %w", resErr, err)

			continue
		}

		_gZaplogSugar.Debugf("Out: %s", spew.Sdump(mOut))
	}

	return resErr
}

/******************************************************************************
 * Functions
 */

func sendRequestToMaster(sender *tele.User) error {
	defer catchPanic()

	msg := fmt.Sprintf("User 'F:%s L:%s U:%s' (ID %d) wants to be invited to %q. "+
		"Do you want to add the user in recipient list?",
		sender.FirstName, sender.LastName, sender.Username, sender.ID, _gTelebotUsername)

	accpetRejectSelector := makeAcceptRejectButtons(sender.ID)

	out, err := _gTelebot.Send(
		masterUser(),
		msg,
		accpetRejectSelector,
	)
	if err != nil {
		_gZaplogSugar.Errorf("Cannot send message. Error: %s", err)

		return err
	}

	_gZaplogSugar.Debug(spew.Sdump(out))

	return nil
}
