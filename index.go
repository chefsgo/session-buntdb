package session_buntdb

import (
	"github.com/chefsgo/chef"
)

func Driver(ss ...string) chef.SessionDriver {
	store := ":memory:"
	if len(ss) > 0 {
		store = ss[0]
	}
	return &buntdbSessionDriver{store}
}

func init() {
	// chef.Register("memory", Driver(":memory:"))
	chef.Register("buntdb", Driver("store/session.db"))
	// chef.Register("file", Driver("store/session.db"))
}
