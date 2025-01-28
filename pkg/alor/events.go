package alor

import "encoding/json"

type Event struct {
	Opcode Opcode
	Guid   string
	Data   json.RawMessage
}
