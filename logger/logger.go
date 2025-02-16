package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

var enabled bool

func init() {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	enabled = false
}

func Init(e bool) {
	enabled = e
}

func PrintLog(chatId int64, message string, err error) {
	if enabled {
		logs := log.Debug().Int64("chat_id", chatId)
		if err != nil {
			logs = logs.Err(err)
		}
		logs.Msg(message)
	}
}
