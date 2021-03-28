package telebot

import (
	"errors"
	"fmt"
	"sync"

	tele "gopkg.in/tucnak/telebot.v3"
)

/******************************************************************************
 * Types
 */

type recipients struct {
	mutex sync.Mutex
	rcpns []tele.Recipient
}

/******************************************************************************
 * Global variables
 */

var (
	ErrorUnknownType = errors.New("unknown type")
)

/******************************************************************************
 * Methods
 */

func (r *recipients) reset() {
	r.mutex.Lock()
	r.rcpns = make([]tele.Recipient, 0)
	r.mutex.Unlock()
}

func (r *recipients) add(rcp tele.Recipient) {
	r.mutex.Lock()
	r.rcpns = append(r.rcpns, rcp)
	r.mutex.Unlock()
}

func (r *recipients) remove(id int) error {
	r.mutex.Lock()

	defer r.mutex.Unlock()

	for i := range r.rcpns {
		switch v := r.rcpns[i].(type) {
		case *tele.User:
			if v.ID == id {
				last := len(r.rcpns) - 1
				r.rcpns[i] = r.rcpns[last]
				r.rcpns[last] = nil
				r.rcpns = r.rcpns[:last]
			}

			return nil

		default:
			return fmt.Errorf("%w %T", ErrorUnknownType, r.rcpns[i])
		}
	}

	return nil
}
