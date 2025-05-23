package tools

import "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"

func MessageTypeToString(messageType int8) string {
	switch messageType {
	case int8(base.InternalMessageType_CHAT_MESSAGE):
		return "Chat"
	case int8(base.InternalMessageType_CONTRAL_MESSAGE):
		return "Control"
	case int8(base.InternalMessageType_QUIZ_CHOICE):
		return "Choice"
	case int8(base.InternalMessageType_QUIZ_JUDGE):
		return "Judge"
	case int8(base.InternalMessageType_QUIZ_STATUS):
		return "Status"
	default:
		return "Unknown"
	}
}
