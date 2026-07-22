package types

import "encoding/json"

type MqMsg struct {
	Uid   uint64          `json:"user_id"`
	ReqId uint64          `json:"req_id"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}
