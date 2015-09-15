package user

import (
	"errors"
	"github.com/boltdb/bolt"

	"github.com/MochiKung/account-interface/config"
	"github.com/MochiKung/account-interface/handler/oauth2/database"
)

const ()

var (
	db *bolt.DB
)

func init() {
	var err error
	db, err = bolt.Open(config.Default.Database.BoltDB.UserDB, 0600, nil)
	if err != nil {
		panic("fail to open database for access-token")
	}
}

type UserInfo struct {
	UID               string
	Username          string
	EncryptedPassword []byte
	Salt              []byte
}

func GetUserInfo(username string) (*UserInfo, error) {
	var userInfo *UserInfo
	err := db.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(username))
		if userBucket == nil {
			return nil
		}
		userInfo = &UserInfo{}
		userInfo.UID = string(userBucket.Get([]byte("uid")))
		userInfo.Username = string(userBucket.Get([]byte("username")))
		userInfo.EncryptedPassword = userBucket.Get([]byte("password"))
		userInfo.Salt = userBucket.Get([]byte("salt"))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func PutUserInfo(userInfo *UserInfo) error {
	oldUserInfo, err := GetUserInfo(userInfo.Username)
	if err != nil && err.Error() != "user not exist" {
		return err
	}
	if oldUserInfo != nil {
		return errors.New("duplicate user")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		userBucket, err := tx.CreateBucket([]byte(userInfo.Username))
		if err != nil {
			return err
		}
		err = database.AddKeyValue(userBucket, "uid", userInfo.UID)
		if err != nil {
			return err
		}
		err = database.AddKeyValue(userBucket, "username", userInfo.Username)
		if err != nil {
			return err
		}
		err = database.AddKeyValue(userBucket, "password", userInfo.EncryptedPassword)
		if err != nil {
			return err
		}
		err = database.AddKeyValue(userBucket, "salt", userInfo.Salt)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
