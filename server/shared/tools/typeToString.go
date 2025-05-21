package tools

import "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"

func MessageTypeToString(messageType int8) string {
	switch messageType {
	case int8(base.InternalMessageType_CHAT_MESSAGE):
		return "Request"
	default:
		return "Unknown"
	}
}
