package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
)

func LoginCellphone(phone, password string) (err error) {
	if phone == "" || password == "" {
		return errors.New("no phone or password")
	}
	md5Bytes := md5.Sum([]byte(password))
	password = hex.EncodeToString(md5Bytes[:])
	debugV("password", password)
	params := map[string]interface{}{
		"phone":       phone,
		"remember":    "true",
		"password":    password,
		"type":        "1",
		"countrycode": "86",
		"e_r":         false,
	}
	data := EncryptedPostData{}
	if err = data.New("/api/login/cellphone", params, true); err == nil {
		_, err = data.DoPost()
	}
	return err
}
