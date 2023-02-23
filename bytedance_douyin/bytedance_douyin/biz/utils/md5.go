package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMD5(password string) string {
	salt := "salt"
	//salt := config.Config.MD5.Salt
	bytes := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(bytes[:])
}
