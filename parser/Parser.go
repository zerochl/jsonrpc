package parser

import (
	"encoding/json"
	"jsonrpc/entity"
	"log"
)

func Parser(jsonStr string) entity.MsgEntity {
	b := []byte(jsonStr)
	var msg entity.MsgEntity
	json.Unmarshal(b, &msg)
	log.Println("result:", msg.URL)
	return msg
}
