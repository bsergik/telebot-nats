package telebot

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tele "gopkg.in/tucnak/telebot.v3"
)

/******************************************************************************
 * Global variables
 */

var (
	acceptInviteCallbackPrefix = fmt.Sprintf("\f%s|", acceptInviteUniqueName)
	rejectInviteCallbackPrefix = fmt.Sprintf("\f%s|", rejectInviteUniqueName)
)

func invitemeHandler(ctx tele.Context) error {
	_gZaplogSugar.Debug(spew.Sdump(ctx))

	defer catchPanic()

	replyMsg := "Request sent to admin."
	sender := ctx.Message().Sender

	err := sendRequestToMaster(sender)
	if err != nil {
		replyMsg = "Cannot sent request to admin. Try again later."
	}

	err = ctx.Reply(replyMsg)
	if err != nil {
		_gZaplogSugar.Errorf("Cannot send message. Error: %s", err)

		return err
	}

	return nil
}

func forgetmeHandler(ctx tele.Context) error {
	_gZaplogSugar.Debug(spew.Sdump(ctx))

	defer catchPanic()

	sender := ctx.Message().Sender

	err := _gRecipients.remove(sender.ID)
	if err != nil {
		_gZaplogSugar.Errorf("Cannot remove recipient. Error: %s", err)

		return err
	}

	err = _gStorage.RemoveRecipient(sender.ID)
	if err != nil {
		_gZaplogSugar.Errorf("Cannot remove recipient. Error: %s", err)

		return err
	}

	if err2 := ctx.Reply("Deleted"); err2 != nil {
		_gZaplogSugar.Errorf("Cannot send message. Error: %s", err2)
	}

	return nil
}

func onCallbackHandler(ctx tele.Context) (err error) {
	_gZaplogSugar.Debug(spew.Sdump(ctx))

	defer catchPanic()

	cb := &tele.CallbackResponse{Text: "success"} //nolint:exhaustivestruct
	data := ctx.Callback().Data

	switch {
	case strings.HasPrefix(data, acceptInviteCallbackPrefix):
		err = acceptInvite(data)

	case strings.HasPrefix(data, rejectInviteCallbackPrefix):
		err = rejectInvite(data)

	default:
		err = fmt.Errorf("not expected callback %q", data)
	}

	if err != nil {
		_gZaplogSugar.Error(err)

		cb.Text = "failed"
	}

	if err2 := ctx.Respond(cb); err2 != nil {
		_gZaplogSugar.Error(err2)
	}

	return err
}
