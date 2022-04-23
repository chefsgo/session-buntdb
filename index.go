package session_buntdb

import (
	"github.com/chefsgo/session"
)

func Driver(ss ...string) session.Driver {
	store := ":memory:"
	if len(ss) > 0 {
		store = ss[0]
	}
	return &buntdbDriver{store}
}

func init() {
	session.Register("buntdb", Driver("store/session.db"))
}
