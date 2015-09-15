package accesstoken

import (
	"dev.corp.extreme.co.th/exe-account/account-interface/config"
	"errors"
	"github.com/boltdb/bolt"
	"time"
)

const ()

var (
	db *bolt.DB
)

func init() {
	var err error
	db, err = bolt.Open(config.Default.Database.BoltDB.AccessTokenDB, 0600, nil)
	if err != nil {
		panic("fail to open database for access-token")
	}
}

type TokenInfo struct {
	Token      string
	Client     string
	User       string
	Scopes     string
	ExpireTime *time.Time
}

func GetTokenInfo(token string, client string) (*TokenInfo, error) {
	var tokenInfo *TokenInfo = nil

	err := db.View(func(tx *bolt.Tx) error {
		clientBucket := tx.Bucket([]byte(client))
		if clientBucket != nil {
			tokenBucket := clientBucket.Bucket([]byte(token))
			if tokenBucket != nil {
				tokenInfo = &TokenInfo{
					Token:  token,
					Client: client,
					User:   queryString(tokenBucket, "user"),
					Scopes: queryString(tokenBucket, "scopes"),
				}

				expireTime := &time.Time{}
				timeBinary := tokenBucket.Get([]byte("expire-time"))
				err := expireTime.UnmarshalBinary(timeBinary)
				if err != nil {
					return err
				}
				tokenInfo.ExpireTime = expireTime
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

func PutTokenInfo(tokenInfo *TokenInfo) error {
	oldTokenInfo, err := GetTokenInfo(tokenInfo.Token, tokenInfo.Client)
	if err != nil {
		return err
	}
	if oldTokenInfo != nil {
		return errors.New("duplicate token")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		var clientBucket *bolt.Bucket
		clientBucket = tx.Bucket([]byte(tokenInfo.Client))
		if clientBucket == nil {
			clientBucket, err = tx.CreateBucket([]byte(tokenInfo.Client))
			if err != nil {
				return err
			}
		}
		tokenBucket, err := clientBucket.CreateBucket([]byte(tokenInfo.Token))
		if err != nil {
			return err
		}
		tokenBucket.Put([]byte("token"), []byte(tokenInfo.Token))
		tokenBucket.Put([]byte("client"), []byte(tokenInfo.Client))
		tokenBucket.Put([]byte("user"), []byte(tokenInfo.User))
		tokenBucket.Put([]byte("scopes"), []byte(tokenInfo.Scopes))
		expireBinary, err := tokenInfo.ExpireTime.MarshalBinary()
		if err != nil {
			return err
		}
		tokenBucket.Put([]byte("expire-time"), expireBinary)
		return nil
	})
	return err
}

func queryString(bucket *bolt.Bucket, key string) string {
	value := bucket.Get([]byte(key))
	if value == nil {
		return ""
	} else {
		return string(value)
	}
}
