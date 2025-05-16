package main

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/util"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

// LectureServiceImpl implements the last service interface defined in the IDL.
type LectureServiceImpl struct {
	MysqlManager
}

type MysqlManager interface {
	CreateLecture(ctx context.Context, lecture *model.Lecture) error
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

// CreateLecture implements the LectureServiceImpl interface.
func (s *LectureServiceImpl) Start(ctx context.Context, request *lecture.StartRequest) (*lecture.StartResponse, error) {
	// ------------ 第一阶段：从主播那里拿到第一个 Offer  ------------
	offer := webrtc.SessionDescription{}
	util.Decode(request.Sdp, &offer)

	// ------------ 配置 PeerConnection 相关  ------------
	// ICE 协商要用的 STUN 服务器（帮助打 NAT 穿透洞）
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	// 创建一个 MediaEngine 对象来配置支持的编解码器
	mediaEngine := &webrtc.MediaEngine{}

	// 注册 VP8 编解码器
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil,
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		klog.Error("Failed to register VP8 codec: ", err)
	}
	// 注册 Opus 编解码器
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil,
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		klog.Error("Failed to register Opus codec: ", err)
	}

	// 拦截器管道，用于处理 NACK、RTCP 报告等
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		klog.Error("Failed to register default interceptors: ", err)
	}
	// 每隔 3 秒发一次 PLI（请求关键帧），保证观众能随机加入时拿到关键帧
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		klog.Error("Failed to create interval PLI factory:", err)
	}
	interceptorRegistry.Add(intervalPliFactory)

	// ------------ 创建广播用的 PeerConnection  ------------
	peerConnection, err := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	).NewPeerConnection(peerConnectionConfig)
	if err != nil {
		klog.Error("Failed to create PeerConnection: ", err)
	}
	defer peerConnection.Close()

	// 允许接收 1 个音频轨道和 1 个视频轨道
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		klog.Error("Failed to add audio transceiver: ", err)
	}
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		klog.Error("Failed to add video transceiver: ", err)
	}
	// 此处，数据库操作。
	sf, err := snowflake.NewNode(consts.LectureIDSnowFlakeNode)
	if err != nil {
		klog.Error("Failed to create snowflake node: ", err)
	}
	err = s.CreateLecture(ctx, &model.Lecture{
		HostId:      request.HostId,
		RoomId:      sf.Generate().Int64(),
		Title:       request.Title,
		Description: request.Description,
		Speaker:     request.Speaker,
		Date:        time.Now(),
	})
	if err != nil {
		klog.Error("Failed to create lecture: ", err)
	}

	// localTrackChan 用来拿到“转发用”的本地Track
	audioLocalTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
	videoLocalTrackChan := make(chan *webrtc.TrackLocalStaticRTP)

	// 当收到主播发过来的远端轨道（OnTrack）时：
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// 1）创建一个本地静态RTP轨道，后面给所有观众用
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.Kind().String(),
			"pion",
		)
		if err != nil {
			klog.Error("Failed to create local track: ", err)
		}
		// 把这个本地轨道传给后续代码
		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			audioLocalTrackChan <- localTrack
		} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
			videoLocalTrackChan <- localTrack
		}

		// 2）不停读主播发来的 RTP 包，并写到本地轨道（广播给所有观众）
		buf := make([]byte, 1400)
		for {
			n, _, readErr := remoteTrack.Read(buf)
			if readErr != nil {
				klog.Error("Failed to read RTP packet: ", readErr)
			}
			// 如果没有观众订阅，Write 会返回 ErrClosedPipe，忽略之
			if _, err := localTrack.Write(buf[:n]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				klog.Error("Failed to write RTP packet: ", err)
			}
		}
	})

	// ------------ 完成与主播的握手  ------------
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		klog.Error("Failed to set remote description: ", err)
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		klog.Error("Failed to create answer: ", err)
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		klog.Error("Failed to set local description: ", err)
	}
	<-gatherComplete // 等 ICE 候选收集完

	resp := &lecture.StartResponse{
		Response: &base.BaseResponse{
			Code:    rsp.Success,
			Message: "success",
		},
		RoomId: 123456,
		Answer: util.Encode(peerConnection.LocalDescription()),
	}
	return resp, nil
}

// Attend implements the LectureServiceImpl interface.
func (s *LectureServiceImpl) Attend(ctx context.Context, request *lecture.AttendRequest) (resp *lecture.AttendResponse, err error) {
	// TODO: Your code here...
	return
}
