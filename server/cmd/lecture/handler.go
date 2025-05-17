package main

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/util"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/pion/webrtc/v4"
)

// LectureServiceImpl implements the last service interface defined in the IDL.
type LectureServiceImpl struct {
	Sessions             *sync.Map
	WebrtcAPI            *webrtc.API
	peerConnectionConfig *webrtc.Configuration

	MysqlManager
}

type MysqlManager interface {
	CreateLecture(ctx context.Context, lecture *model.Lecture) error
	RecordJoin(ctx context.Context, Attendance *model.Attendance) error
	RecordLeft(ctx context.Context, id int64) error
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

// 房间号对应的 LectureSession 结构体
// 每个结构体代表房间里面主播的轨道以及主播的 PeerConnection 以及观众的连接状况
type LectureSession struct {
	PeerConnection *webrtc.PeerConnection
	AudioTrack     chan *webrtc.TrackLocalStaticRTP
	VideoTrack     chan *webrtc.TrackLocalStaticRTP
	Audiences      *sync.Map
}

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
	// ------------ 创建广播用的 PeerConnection  ------------
	peerConnection, err := s.WebrtcAPI.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		klog.Error("Failed to create PeerConnection: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorCreatePeerConnection(err.Error()),
		}, nil
	}

	// 允许接收 1 个音频轨道和 1 个视频轨道
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		klog.Error("Failed to add audio transceiver: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorAddtTransceiver(err.Error()),
		}, nil
	}
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		klog.Error("Failed to add video transceiver: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorAddtTransceiver(err.Error()),
		}, nil
	}
	// 此处，数据库操作。
	// 雪花算法生成唯一的房间号，作为 id
	sf, err := snowflake.NewNode(consts.LectureIDSnowFlakeNode)
	if err != nil {
		klog.Error("Failed to create snowflake node: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorSnowFalke(err.Error()),
		}, nil
	}
	roomid := sf.Generate().Int64()
	err = s.CreateLecture(ctx, &model.Lecture{
		HostId:      request.HostId,
		RoomId:      roomid,
		Title:       request.Title,
		Description: request.Description,
		Speaker:     request.Speaker,
		Date:        time.Now(),
	})
	if err != nil {
		klog.Error("Failed to create lecture: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorCreateLecture(err.Error()),
		}, nil
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
			return
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
				return
			}
			// 如果没有观众订阅，Write 会返回 ErrClosedPipe，忽略之
			if _, err := localTrack.Write(buf[:n]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				klog.Error("Failed to write RTP packet: ", err)
				return
			}
		}
	})

	// ------------ 完成与主播的握手  ------------

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		klog.Error("Failed to set remote description: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorSetRemoteDescription(err.Error()),
		}, nil
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		klog.Error("Failed to create answer: ", err)
		return &lecture.StartResponse{
			Response: rsp.ErrorCreateAnswer(err.Error()),
		}, nil
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		klog.Error("Failed to set local description: ", err)

	}
	<-gatherComplete // 等 ICE 候选收集完
	s.Sessions.Store(roomid, &LectureSession{
		PeerConnection: peerConnection,
		AudioTrack:     audioLocalTrackChan,
		VideoTrack:     videoLocalTrackChan,
		Audiences:      &sync.Map{},
	})
	return &lecture.StartResponse{
		Response: rsp.OK(),
		RoomId:   roomid,
		Answer:   util.Encode(peerConnection.LocalDescription()),
	}, nil
}

// Attend implements the LectureServiceImpl interface.
// 学生出席课程的逻辑部分，首先通过房间号获取讲师的轨道，然后和学生的 PeerConnection 做一次 Offer/Answer + ICE 协商，
// 最后把 Answer 返回给前端处理。
func (s *LectureServiceImpl) Attend(ctx context.Context, request *lecture.AttendRequest) (*lecture.AttendResponse, error) {
	// 收到下一个观众发来的 Offer（Base64）
	recvOnlyOffer := webrtc.SessionDescription{}
	util.Decode(request.Sdp, &recvOnlyOffer)

	// 为这个观众新建一个 PeerConnection
	pc, err := s.WebrtcAPI.NewPeerConnection(*s.peerConnectionConfig)
	if err != nil {
		klog.Error("Failed to create PeerConnection: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorCreatePeerConnection(err.Error()),
		}, nil
	}
	// 从哈希表中获取对应房间的 LectureSession
	session, ok := s.Sessions.Load(request.RoomId)
	if !ok {
		klog.Error("Failed to find session: ", request.RoomId)
		return &lecture.AttendResponse{
			Response: rsp.ErrorRoomNotFound(),
		}, nil
	}
	Session := session.(*LectureSession)
	// 存储这个 PeerConnection 到 LectureSession 的 Audiences 哈希表中
	// 便于后续可以扩展。
	Session.Audiences.Store(request.UserId, pc)
	sf, err := snowflake.NewNode(consts.AttendanceIDSnowFlakeNode)
	AttendanceId := sf.Generate().Int64()
	s.RecordJoin(ctx, &model.Attendance{
		AttendanceId: AttendanceId,
		RoomId:       request.RoomId,
		UserId:       request.UserId,
	})
	// 监听 PeerConnection 的 ICE 连接状态
	// 如果学生关闭网页，或者网络断开，就关闭这个 PeerConnection
	// 并且记录退出时间，便于统计。
	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected || connectionState == webrtc.ICEConnectionStateFailed {
			s.RecordLeft(ctx, AttendanceId)
			Session.Audiences.Delete(request.UserId)
			pc.Close()
		}
	})
	AudioRtcp, err := pc.AddTrack(<-Session.AudioTrack)
	if err != nil {
		klog.Error("Failed to add audio track: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorAddTrack(err.Error()),
		}, nil
	}

	VideoRtcp, err := pc.AddTrack(<-Session.VideoTrack)
	if err != nil {
		klog.Error("Failed to add video track: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorAddTrack(err.Error()),
		}, nil
	}
	// 启动协程，也许可以用协程池？
	// 需要读 RTCP 包触发 NACK/PLI 等功能
	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := AudioRtcp.Read(buf); rtcpErr != nil {
				return
			}
		}
	}()
	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := VideoRtcp.Read(buf); rtcpErr != nil {
				return
			}
		}
	}()

	// 和这个观众做一次 Offer/Answer + ICE
	if err = pc.SetRemoteDescription(recvOnlyOffer); err != nil {
		klog.Error("Failed to set remote description: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorSetRemoteDescription(err.Error()),
		}, nil
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		klog.Error("Failed to create answer: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorCreateAnswer(err.Error()),
		}, nil
	}
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(answer); err != nil {
		klog.Error("Failed to set local description: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorSetLocalDescription(err.Error()),
		}, nil
	}
	<-gatherComplete

	klog.Info("Audience ICE gathering complete")
	return &lecture.AttendResponse{
		Response: rsp.OK(),
		Answer:   util.Encode(pc.LocalDescription()),
	}, nil
}
