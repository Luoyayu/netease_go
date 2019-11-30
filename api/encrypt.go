package api

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/hex"
)

// EncryptData :加密数据
// key 16位密钥
// data 待加密文本

const (
	magicKey = "e82ckenh8dichen8"
)

type EncryptData struct {
	key  []byte
	data []byte
}

// New :新建待加密类
func (r *EncryptData) New(key, data []byte) {
	r.key = key
	r.data = data
}

func (r *EncryptData) NewWithDefaultKey(data []byte) {
	r.key = []byte(magicKey)
	r.data = data

}

// EcbEncrypt :AES ECB 模式 采用PKCS#7 填充
func (r *EncryptData) EcbEncrypt() ([]byte, error) {
	if b, err := aes.NewCipher(r.key); err == nil {
		bs := b.BlockSize()
		paddingData := _PKCS7Padding(r.data, bs)
		decrypted := make([]byte, len(paddingData))
		for s, e := 0, bs; s < len(paddingData); s, e = s+bs, e+bs {
			b.Encrypt(decrypted[s:e], paddingData[s:e])
		}
		return decrypted, nil
	} else {
		return nil, err
	}
}

//  EcbDecrypt :AES ECB 模式 采用PKCS#7 解填充
func (r *EncryptData) EcbDecrypt() ([]byte, error) {
	if b, err := aes.NewCipher(r.key); err == nil {
		bs := b.BlockSize()

		decrypted := make([]byte, len(r.data))
		for s, e := 0, bs; s < len(r.data); s, e = s+bs, e+bs {
			b.Decrypt(decrypted[s:e], r.data[s:e])
		}
		return _PKCS7UnPadding(decrypted), nil
	} else {
		return nil, err
	}
}

// _PKCS7Padding :PKCS#7 填充
func _PKCS7Padding(ciphertext []byte, bs int) []byte {
	padding := bs - len(ciphertext)%bs
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// _PKCS7UnPadding :PKCS#7 解填充
func _PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// MD5ForEncrypt :md5 加密路径和字典, 生成sign
func MD5ForEncrypt(path, dict string) string {
	st := "nobody" + path + "use" + dict + "md5forencrypt"
	md5Bytes := md5.Sum([]byte(st))
	return hex.EncodeToString(md5Bytes[:])
}
