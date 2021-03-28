package telebot

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/bsergik/telebot-nats/internal/database"
	"github.com/davecgh/go-spew/spew"
	tele "gopkg.in/tucnak/telebot.v3"
)

func catchPanic() {
	if rcv := recover(); rcv != nil {
		_gZaplogSugar.Errorf("cought panic!\nError: %s;\nTrace: %s", rcv, string(debug.Stack()))
	}
}

func fillInRecipientsFromDB() {
	rcpns, err := _gStorage.GetRecipients()
	if err != nil {
		_gZaplogSugar.Errorf("Cannot get recipients from DB. Error: %s", err)

		return
	}

	for i := range rcpns {
		_gRecipients.add(&tele.User{ID: int(rcpns[i].ID)})
	}

	_gZaplogSugar.Debug(spew.Sprintf("%#v", rcpns))
}

func masterUser() *tele.User {
	return &tele.User{ID: _gMasterUserID}
}

func simpleUser(id int) *tele.User {
	return &tele.User{ID: id}
}

func formatMessage(m *database.Message) string {
	return fmt.Sprintf("*%s* at _%s_\n```\n%s\n```\n",
		strings.ToUpper(m.Subsystem), m.CreatedAt, m.Message)
}

func getIDfromInvite(data string) (int64, error) {
	d := strings.Split(data, "|")
	if len(d) == 2 {
		return strconv.ParseInt(d[1], 10, 64)
	}

	return 0, nil
}

func acceptInvite(data string) error {
	id, err := getIDfromInvite(data)
	if err != nil || id == 0 {
		err := fmt.Errorf("acceptInvite: not expected ID %d. Error: %v", id, err)

		return err
	}

	_gRecipients.add(simpleUser(int(id)))

	err = _gStorage.AddRecipient(int(id))
	if err != nil {
		err := fmt.Errorf("acceptInvite: cannot save recipient %d. Error: %v", id, err)

		return err
	}

	user := &tele.User{ID: int(id)}

	_, err = _gTelebot.Send(user, "Your request accepted.\nPlease Unmute the bot in order to receive urgent messages.")
	if err != nil {
		err := fmt.Errorf("acceptInvite: cannot send message to recipient %d. Error: %v", id, err)

		return err
	}

	return nil
}

func rejectInvite(data string) error {
	id, err := getIDfromInvite(data)
	if err != nil || id == 0 {
		err := fmt.Errorf("rejectInvite: not expected ID %d. Error: %v", id, err)

		return err
	}

	err = _gRecipients.remove(int(id))
	if err != nil {
		err := fmt.Errorf("rejectInvite: cannot remove %d. Error: %v", id, err)

		return err
	}

	err = _gStorage.RemoveRecipient(int(id))
	if err != nil {
		err := fmt.Errorf("rejectInvite: cannot save recipient %d. Error: %v", id, err)

		return err
	}

	return nil
}

func makeAcceptRejectButtons(senderID int) *tele.ReplyMarkup {
	senderIDstr := fmt.Sprintf("%d", senderID)
	selector := new(tele.ReplyMarkup)
	btnAccept := selector.Data(acceptButton, acceptInviteUniqueName, senderIDstr)
	btnReject := selector.Data(rejectButton, rejectInviteUniqueName, senderIDstr)

	selector.Inline(
		selector.Row(btnAccept, btnReject),
	)

	return selector
}
