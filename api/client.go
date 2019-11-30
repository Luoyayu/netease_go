package api

/*
# example
```
	params := map[string]interface{}{
		"verifyId": 1, "id": "514774419", "tv": "-1", "lv": "-1", "kv": "-1", "e_r": false,
	}
	data := api.EncryptedPostData{}
	_ = data.New("/api/song/lyric", params, false) // false -> use MUSIC_U; true -> MUSIC_A
	if ret, err := data.DoPost(); err == nil {
		log.Println(string(ret))
	} else {
		log.Println(err)
	}
```
*/

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	magicJoin = "-36cd479b6b5-"
	DEBUG     = false
)

func debugV(name string, v ...interface{}) {
	if DEBUG {
		fmt.Println(name, ": ", v)
	}
}

var HeaderWithTokenStr, HeaderStr string

func init() {
	if os.Getenv("DEBUG") != "" {
		DEBUG = true
	}

	cookie, err := os.Open("cookies")
	var headerBytes []byte
	if err != nil {
		header := map[string]interface{}{
			"header": map[string]string{
				"os":     "osx",
				"appver": "2.3.0",
			},
		}
		headerBytes, _ = json.Marshal(header)
		HeaderStr = string(headerBytes)
	} else {
		MUSICUBytes, _ := ioutil.ReadAll(cookie)
		header := map[string]interface{}{
			"header": map[string]string{
				"MUSIC_U": string(MUSICUBytes),
				"os":      "osx",
				"appver":  "2.3.0",
			},
		}
		headerBytes, _ = json.Marshal(header)
		HeaderWithTokenStr = string(headerBytes)
	}
}

// NetEaseClient :网易云HTTP客户端
type NetEaseClient struct {
	c *http.Client
	q *http.Request
	p *http.Response
	d string // post body
}

// NewDefaultClient :新建默认网易云HTTP客户端
func (r *NetEaseClient) NewDefaultClient() {
	r.c = &http.Client{}
}

// post :私有, 仅对加密的body有效
func (r *NetEaseClient) post(url string) (b []byte, err error) {
	if r.d == "" {
		return nil, errors.New("no body data as params")
	}
	if r.q, err = http.NewRequest("POST", url, bytes.NewBuffer([]byte(r.d))); err == nil {
		r.q.Header.Set("host", "music.163.com")
		r.q.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko)")
		r.q.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.q.Header.Set("cookies", "channel=netease; os=osx")

		if r.p, err = r.c.Do(r.q); err == nil {
			debugV("Header", r.p.Header)
			cookies := r.p.Cookies()
			if cookies != nil {
				for _, c := range cookies {
					if c.Name == "MUSIC_U" {
						log.Println("login success!")
						var f afero.File
						if exist, _ := afero.Exists(appFs, "cookies"); exist == true {
							f, _ = appFs.Open("cookies")
						} else {
							f, _ = appFs.Create("cookies")
						}
						_, _ = f.WriteString(c.Value)
						break
					}
				}
			}
			if b, err = ioutil.ReadAll(r.p.Body); err == nil {

			}
		}
	}
	return
}

// DoPost :加密r.path和r.dict构成的r.text, 发送POST请求
func (r *EncryptedPostData) DoPost() (ret []byte, err error) {
	c := NetEaseClient{}
	c.NewDefaultClient()

	var encryptDataHexStr string
	if encryptDataHexStr, err = r.Encrypt(); err == nil {
		c.d = "params=" + encryptDataHexStr
		debugV("encryptDataHexStr with params", c.d)
		if ret, err = c.post("http://music.163.com/e" + r.path[1:]); err == nil {
			debugV("POST Response Body", string(ret))
			retCode := struct {
				Code int64  `json:"code"`
				Msg  string `json:"msg"`
			}{}
			_ = json.Unmarshal(ret, &retCode)
			if retCode.Code != 200 {
				return nil, errors.New(retCode.Msg)
			}
		}
	}
	return
}

// EncryptedPostData :加密POST的数据
type EncryptedPostData struct {
	path string
	dict string
	text string // Encrypted^Post^Data
}

func (r *EncryptedPostData) New(path string, params map[string]interface{}, firstLogin bool) (err error) {
	r.path = path
	if paramsStr, err := json.Marshal(params); err == nil {
		if firstLogin == false {
			r.dict = strings.TrimSuffix(HeaderWithTokenStr, "}") + "," + strings.TrimPrefix(string(paramsStr), "{")
		} else {
			debugV("HeaderStr", HeaderStr)
			r.dict = strings.TrimSuffix(HeaderStr, "}") + "," + strings.TrimPrefix(string(paramsStr), "{")
		}
		md5hexSign := MD5ForEncrypt(r.path, r.dict)
		r.text = strings.Join([]string{r.path, r.dict, md5hexSign}, magicJoin)
		if DEBUG {
			debugV("dict", r.dict)
			debugV("sign", md5hexSign)
			debugV("text", r.text)
		}
	}
	return
}

// Decrypt : Decrypt and apply to r.path, r.dict
func (r *EncryptedPostData) Decrypt() (err error) {
	var decryptBytes []byte
	if decryptBytes, err = DecryptParams(r.text); err == nil {
		decryptStr := string(decryptBytes)
		decryptList := strings.Split(decryptStr, magicJoin)
		if len(decryptList) < 3 {
			return errors.New("decrypt failed")
		}
		r.path, r.dict = decryptList[0], decryptList[1]
	}
	return
}

// Encrypt Encrypt r.text by EcbEncrypt; return Encrypted Hex String
func (r *EncryptedPostData) Encrypt() (ret string, err error) {
	ret, err = EncryptParams(r.text)
	return
}

func EncryptParams(text string) (ret string, err error) {
	if text == "" {
		err = errors.New("text is None")
		return
	}
	ecb := EncryptData{}
	ecb.NewWithDefaultKey([]byte(text))
	var encryptData []byte
	if encryptData, err = ecb.EcbEncrypt(); err == nil {
		ret = strings.ToUpper(hex.EncodeToString(encryptData))
	}
	return
}

func DecryptParams(text string) (ret []byte, err error) {
	if text == "" {
		err = errors.New("text is None")
		return
	}
	var textBytes []byte
	if textBytes, err = hex.DecodeString(text); err == nil {
		de := EncryptData{}
		de.NewWithDefaultKey(textBytes)
		if ret, err = de.EcbDecrypt(); err == nil {
			debugV("decrypt params: ", string(ret))
		}
	}
	return
}
