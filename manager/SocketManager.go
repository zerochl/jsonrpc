package manager

import (
	"encoding/json"
	"io/ioutil"
	"jsonrpc/parser"
	"log"
	"net/http"
)

func GetTextByUrl(url, cookie, userAgent string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var textByte []byte
	textByte, _ = ioutil.ReadAll(resp.Body)
	log.Println("result:", string(textByte[0:20]))
	return string(textByte)
}

func GetTextByJson(jsonStr string) string {
	msg := parser.Parser(jsonStr)
	msg.HTML_TEXT = GetTextByUrl(msg.URL, msg.COOKIE, msg.USER_AGENT)
	replyJsonStr, err := json.Marshal(msg)
	if err != nil {
		return "err" + err.Error()
	}
	return string(replyJsonStr)
}
