package util

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/pion/webrtc/v4"
)

func Encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		klog.Error("Encode error: ", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

func Decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		klog.Error("Decode error: ", err)
		return
	}
	if err = json.Unmarshal(b, obj); err != nil {
		klog.Error("Decode error: ", err)
		return
	}
}
