package types

import (
	"github.com/open-policy-agent/opa/v1/util"
)

type Message uint8

const (
	ErrorMessage Message = iota + 1
	WarningMessage
	InfoMessage
	LogMessage
)

func (m Message) AppendText(buf []byte) []byte {
	return util.AppendInt(buf, m)
}
