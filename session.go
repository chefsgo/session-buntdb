package session_buntdb

import (
	"errors"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/chefsgo/base"
	"github.com/chefsgo/chef"
	"github.com/chefsgo/session"
	"github.com/tidwall/buntdb"
)

var (
	errInvalidStore    = errors.New("Invalid store.")
	errInvalidDatabase = errors.New("Invalid database.")
)

type (
	buntdbDriver struct {
		store string
	}
	buntdbConnect struct {
		mutex sync.RWMutex

		name    string
		config  session.Config
		setting buntdbSetting

		db *buntdb.DB
	}
	buntdbSetting struct {
		Store  string
		Expiry time.Duration
	}
	buntdbValue struct {
		Value Any `json:"value"`
	}
)

//连接
func (driver *buntdbDriver) Connect(name string, config session.Config) (session.Connect, error) {
	//获取配置信息
	setting := buntdbSetting{
		Store: driver.store,
	}

	if vv, ok := config.Setting["file"].(string); ok && vv != "" {
		setting.Store = vv
	} else if vv, ok := config.Setting["store"].(string); ok && vv != "" {
		setting.Store = vv
	} else {
		setting.Store = "store/session.db"
	}

	dir := path.Dir(setting.Store)
	_, e := os.Stat(dir)
	if e != nil {
		//创建目录，如果不存在
		os.MkdirAll(dir, 0700)
	}

	return &buntdbConnect{
		name: name, config: config, setting: setting,
	}, nil
}

//打开连接
func (connect *buntdbConnect) Open() error {
	if connect.setting.Store == "" {
		return errInvalidStore
	}
	db, err := buntdb.Open(connect.setting.Store)
	if err != nil {
		return err
	}
	connect.db = db
	return nil
}

//关闭连接
func (connect *buntdbConnect) Close() error {
	if connect.db != nil {
		if err := connect.db.Close(); err != nil {
			return err
		}
	}
	return nil
}

//查询缓存，
func (connect *buntdbConnect) Read(key string) (Map, error) {
	if connect.db == nil {
		return nil, errInvalidDatabase
	}

	realVal := ""

	err := connect.db.View(func(tx *buntdb.Tx) error {
		vvv, err := tx.Get(key)
		if err != nil {
			return err
		}
		realVal = vvv
		return nil
	})
	if err != nil {
		return nil, err
	}

	value := Map{}
	err = chef.UnmarshalJSON([]byte(realVal), &value)
	if err != nil {
		return nil, nil
	}

	return value, nil
}

//更新缓存
func (connect *buntdbConnect) Write(key string, val Map, expiry time.Duration) error {
	if connect.db == nil {
		return errInvalidDatabase
	}

	bytes, err := chef.MarshalJSON(val)
	if err != nil {
		return err
	}

	realVal := string(bytes)

	if expiry <= 0 {
		expiry = connect.config.Expiry
	}

	return connect.db.Update(func(tx *buntdb.Tx) error {
		opts := &buntdb.SetOptions{Expires: true, TTL: expiry}
		_, _, err := tx.Set(key, realVal, opts)
		return err
	})
}

//删除缓存
func (connect *buntdbConnect) Delete(key string) error {
	if connect.db == nil {
		return errInvalidDatabase
	}

	return connect.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		return err
	})
}

func (connect *buntdbConnect) Clear(prefix string) error {
	if connect.db == nil {
		return errInvalidDatabase
	}

	return connect.db.Update(func(tx *buntdb.Tx) error {
		keys := make([]string, 0)
		err := tx.AscendKeys(prefix, func(key, value string) bool {
			keys = append(keys, key)
			return true
		})
		if err != nil {
			return err
		}

		for _, key := range keys {
			_, err := tx.Delete(key)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
