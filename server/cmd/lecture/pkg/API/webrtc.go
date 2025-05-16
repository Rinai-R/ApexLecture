package api

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

func NewWebrtcAPI() *webrtc.API {
	// 设置多路复用 UDP 端口
	settingEngine := webrtc.SettingEngine{}
	mux, err := ice.NewMultiUDPMuxFromPort(8443)
	if err != nil {
		klog.Fatal("Failed to create multi-UDP mux: ", err)
	}
	settingEngine.SetICEUDPMux(mux)

	mediaEngine := &webrtc.MediaEngine{}

	// 注册 VP8 编解码器
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil,
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		klog.Fatal("Failed to register VP8 codec: ", err)
	}
	// 注册 Opus 编解码器
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil,
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		klog.Fatal("Failed to register Opus codec: ", err)
	}

	// 拦截器管道，用于处理 NACK、RTCP 报告等
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		klog.Fatal("Failed to register default interceptors: ", err)
	}
	// 每隔 3 秒发一次 PLI（请求关键帧），保证观众能随机加入时拿到关键帧
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		klog.Fatal("Failed to create interval PLI factory:", err)
	}
	interceptorRegistry.Add(intervalPliFactory)

	// ------------ 创建广播用的 PeerConnection  ------------
	API := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
		webrtc.WithSettingEngine(settingEngine),
	)
	return API
}
