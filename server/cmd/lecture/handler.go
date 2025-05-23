package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/util"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/minio/minio-go/v7"
	"github.com/panjf2000/ants/v2"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
	"github.com/pion/webrtc/v4/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v4/pkg/media/oggreader"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
)

// LectureServiceImpl implements the last service interface defined in the IDL.
type LectureServiceImpl struct {
	Sessions             *sync.Map
	WebrtcAPI            *webrtc.API
	peerConnectionConfig *webrtc.Configuration
	goroutinePool        *ants.Pool // 控制并发数
	MinioManager         *minio.Client
	MysqlManager
	RedisManager
}

type MysqlManager interface {
	CreateLecture(ctx context.Context, lecture *model.Lecture) error
	RecordJoin(ctx context.Context, Attendance *model.Attendance) error
	RecordLeft(ctx context.Context, id int64) error
	IsRecorded(ctx context.Context, roomId int64) error
	CheckRecord(ctx context.Context, roomId int64) (bool, error)
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

type RedisManager interface {
	CreateRoom(ctx context.Context, roomId int64, hostid int64) error
	DeleteRoom(ctx context.Context, roomId int64) error
	DeleteSignal(ctx context.Context, roomId int64) error
	AddRoomPerson(ctx context.Context, roomId int64, userId int64) error
	SubRoomPerson(ctx context.Context, roomId int64, userId int64) error
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

// 房间号对应的 LectureSession 结构体
// 每个结构体代表房间里面主播的轨道以及主播的 PeerConnection 以及观众的连接状况
type LectureSession struct {
	HostId          int64
	PeerConnection  *webrtc.PeerConnection
	AudioTrack      *webrtc.TrackLocalStaticRTP
	VideoTrack      *webrtc.TrackLocalStaticRTP
	Audiences       *sync.Map
	AudioRecordChan chan *rtp.Packet
	VideoRecordChan chan *rtp.Packet
	RecordStarted   bool
}

// CreateLecture implements the LectureServiceImpl interface.
func (s *LectureServiceImpl) Start(ctx context.Context, request *lecture.StartRequest) (*lecture.StartResponse, error) {
	// 从主播那里拿到 Offer
	offer := webrtc.SessionDescription{}
	util.Decode(request.Offer, &offer)

	// 配置 PeerConnection 相关
	// ICE 协商要用的 STUN 服务器（帮助打 NAT 穿透洞）
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	// 创建主播用的 PeerConnection
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

	// redis 存储房间号，使得消息服务可以知道这个房间的存在。
	s.RedisManager.CreateRoom(ctx, roomid, request.HostId)

	// 远程调用，创建交互房间。

	// localTrackChan 用来拿到“转发用”的本地Track
	// localTrackChan 用来拿到“转发用”的本地Track
	audioLocalTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
	videoLocalTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
	// 这两个管道用于同步，保证两个轨道都准备好了。
	Check := make(chan struct{}, 2)
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
			<-Check
			session, ok := s.Sessions.Load(roomid)
			if !ok {
				klog.Error("Failed to find session: ", roomid)
				return
			}
			Session := session.(*LectureSession)
			for {
				rtp, _, err := remoteTrack.ReadRTP()
				if err != nil {
					klog.Error("Failed to read RTP packet: ", err)
					return
				}
				if Session.RecordStarted {
					Session.AudioRecordChan <- rtp
				}

				// 完成与主播的握手
				// 如果没有观众订阅，Write 会返回 ErrClosedPipe，忽略之
				if err := localTrack.WriteRTP(rtp); err != nil && !errors.Is(err, io.ErrClosedPipe) {
					klog.Error("Failed to write RTP packet: ", err)
					return
				}
			}
		} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
			videoLocalTrackChan <- localTrack
			<-Check
			session, ok := s.Sessions.Load(roomid)
			if !ok {
				klog.Error("Failed to find session: ", roomid)
				return
			}
			Session := session.(*LectureSession)
			for {
				rtp, _, err := remoteTrack.ReadRTP()
				if err != nil {
					klog.Error("Failed to read RTP packet: ", err)
					return
				}
				if Session.RecordStarted {
					Session.VideoRecordChan <- rtp
				}

				// 完成与主播的握手
				// 如果没有观众订阅，Write 会返回 ErrClosedPipe，忽略之
				if err := localTrack.WriteRTP(rtp); err != nil && !errors.Is(err, io.ErrClosedPipe) {
					klog.Error("Failed to write RTP packet: ", err)
					return
				}
			}
		}
	})
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {
			klog.Info("主播连接断开")
			// 此处也需要告诉其他服务，这个房间已经关闭了
			// 使得消息服务可以正常运作。
			s.RedisManager.DeleteRoom(ctx, roomid)
			s.RedisManager.DeleteSignal(ctx, roomid)
			s.Sessions.Delete(roomid)
			peerConnection.Close()
		}
	})

	// 完成与主播的握手
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
	<-gatherComplete
	// 等 ICE 候选收集完
	// 存储转发音频轨道和视频轨道的管道，便于用户来的时候获取音视频的轨道。
	s.goroutinePool.Submit(func() {
		s.Sessions.Store(roomid, &LectureSession{
			HostId:          request.HostId,
			PeerConnection:  peerConnection,
			AudioTrack:      <-audioLocalTrackChan,
			VideoTrack:      <-videoLocalTrackChan,
			Audiences:       &sync.Map{},
			VideoRecordChan: make(chan *rtp.Packet, 200),
			AudioRecordChan: make(chan *rtp.Packet, 200),
			RecordStarted:   false,
		})
		Check <- struct{}{}
		Check <- struct{}{}
		close(Check)
	})
	return &lecture.StartResponse{
		Response: rsp.OK(),
		RoomId:   roomid,
		Answer:   util.Encode(peerConnection.LocalDescription()),
	}, nil
}

// Attend implements the LectureServiceImpl interface.
// 学生出席课程的逻辑部分，首先通过房间号获取讲师的轨道，
// 然后和学生的 PeerConnection 做一次 Offer/Answer + ICE 协商，
// 最后把 Answer 返回给前端处理。
func (s *LectureServiceImpl) Attend(ctx context.Context, request *lecture.AttendRequest) (*lecture.AttendResponse, error) {
	// 收到观众发来的 Offer（Base64）
	recvOnlyOffer := webrtc.SessionDescription{}
	util.Decode(request.Offer, &recvOnlyOffer)

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
			Response: rsp.ErrorRoomNotExists("Room Not Found"),
		}, nil
	}
	Session := session.(*LectureSession)
	// 存储这个 PeerConnection 到 LectureSession 的 Audiences 哈希表中
	// 便于后续可以扩展。
	Session.Audiences.Store(request.UserId, pc)
	sf, err := snowflake.NewNode(consts.AttendanceIDSnowFlakeNode)
	if err != nil {
		klog.Error("Failed to Create snowflake NewNode", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorSnowFalke(err.Error()),
		}, nil
	}
	AttendanceId := sf.Generate().Int64()
	s.RecordJoin(ctx, &model.Attendance{
		AttendanceId: AttendanceId,
		RoomId:       request.RoomId,
		UserId:       request.UserId,
	})
	// 增加 redis 中记录的房间人数，让其他服务可以实时获取人数，比如答题状态统计和推送
	s.RedisManager.AddRoomPerson(ctx, request.RoomId, request.UserId)
	// 监听 PeerConnection 的 ICE 连接状态
	// 如果学生关闭网页，或者网络断开，就关闭这个 PeerConnection
	// 并且记录退出时间，如果后续扩展，可以便于统计。
	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected || connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateClosed {
			// 记录退出时间
			// 并且清理状态
			s.RecordLeft(ctx, AttendanceId)
			Session.Audiences.Delete(request.UserId)
			s.RedisManager.SubRoomPerson(ctx, request.RoomId, request.UserId)
			pc.Close()
		}
	})
	AudioRtcp, err := pc.AddTrack(Session.AudioTrack)
	if err != nil {
		klog.Error("Failed to add audio track: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorAddTrack(err.Error()),
		}, nil
	}

	VideoRtcp, err := pc.AddTrack(Session.VideoTrack)
	if err != nil {
		klog.Error("Failed to add video track: ", err)
		return &lecture.AttendResponse{
			Response: rsp.ErrorAddTrack(err.Error()),
		}, nil
	}

	// 启动协程，也许可以用协程池？
	// 需要读 RTCP 包触发 NACK/PLI 等功能
	s.goroutinePool.Submit(func() {
		buf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := AudioRtcp.Read(buf); rtcpErr != nil {
				return
			}
		}
	})
	s.goroutinePool.Submit(func() {
		buf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := VideoRtcp.Read(buf); rtcpErr != nil {
				return
			}
		}
	})

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

// Record implements the LectureServiceImpl interface.
// Record 录制视频，接收 RTP 包，写入临时文件，然后上传到 Minio。
func (s *LectureServiceImpl) Record(ctx context.Context, request *lecture.RecordRequest) (*lecture.RecordResponse, error) {
	session, ok := s.Sessions.Load(request.RoomId)
	if !ok {
		klog.Error("Failed to find session: ", request.RoomId)
		return &lecture.RecordResponse{
			Response: rsp.ErrorRoomNotExists("Room Not Found"),
		}, nil
	}
	Session := session.(*LectureSession)
	if Session.RecordStarted {
		return &lecture.RecordResponse{
			Response: rsp.ErrorRecordAlreadyStarted(),
		}, nil
	}
	Session.RecordStarted = true
	// 先把临时文件准备好。
	OggFile := fmt.Sprintf(consts.OggFilePath, request.RoomId)
	IvfFile := fmt.Sprintf(consts.IvfFilePath, request.RoomId)
	err := os.MkdirAll(filepath.Dir(OggFile), 0777)
	if err != nil {
		klog.Error("Failed to create directory: ", err)
		return &lecture.RecordResponse{
			Response: rsp.ErrorSaveLecture(err.Error()),
		}, nil
	}
	err = os.MkdirAll(filepath.Dir(IvfFile), 0777)
	if err != nil {
		klog.Error("Failed to create directory: ", err)
		return &lecture.RecordResponse{
			Response: rsp.ErrorSaveLecture(err.Error()),
		}, nil
	}
	f, err := os.Create(OggFile)
	if err != nil {
		klog.Error("Failed to create file: ", err)
		return &lecture.RecordResponse{
			Response: rsp.ErrorSaveLecture(err.Error()),
		}, nil
	}
	f.Close()
	f, err = os.Create(IvfFile)
	if err != nil {
		klog.Error("Failed to create file: ", err)
		return &lecture.RecordResponse{
			Response: rsp.ErrorSaveLecture(err.Error()),
		}, nil
	}
	f.Close()
	// 临时文件准备完成。
	oggFile, err := oggwriter.New(OggFile, 48000, 2)
	if err != nil {
		panic(err)
	}
	ivfFile, err := ivfwriter.New(IvfFile, ivfwriter.WithCodec("video/VP8"))
	if err != nil {
		panic(err)
	}
	// 异步保存，不阻塞。
	s.goroutinePool.Submit(func() {
		s.Save(ctx, OggFile, oggFile, Session.AudioRecordChan)
	})
	s.goroutinePool.Submit(func() {
		s.Save(ctx, IvfFile, ivfFile, Session.VideoRecordChan)
	})

	return &lecture.RecordResponse{
		Response: rsp.OK(),
	}, nil
}

// 通过临时文件来异步保存的逻辑
func (s *LectureServiceImpl) Save(ctx context.Context, filepath string, writer media.Writer, ch chan *rtp.Packet) {
	for {
		select {
		case packet := <-ch:
			err := writer.WriteRTP(packet)
			if err != nil {
				klog.Error("Record: Failed to write RTP packet: ", err)
			}
		case <-time.Tick(time.Second * 10):
			goto minio
		}
	}
minio:
	err := writer.Close()
	if err != nil {
		klog.Error("Record: Failed to close writer: ", err)
		return
	}
	// 保存文件。
	f, err := os.Open(filepath)
	if err != nil {
		klog.Error("Record: Failed to open file: ", err)
		return
	}
	fileInfo, err := f.Stat()
	if err != nil {
		klog.Error("Record: Failed to get file info: ", err)
		return
	}
	defer f.Close()
	var contentType string
	if strings.HasSuffix(filepath, ".ivf") {
		contentType = "video/ivf"
	}
	if strings.HasSuffix(filepath, ".ogg") {
		contentType = "audio/ogg"
	}
	dir := strings.Split(filepath, "/")
	roomId, err := strconv.ParseInt(dir[len(dir)-2], 10, 64)
	if err != nil {
		klog.Error("Record: Failed to parse room id: ", err)
		return
	}
	_, err = s.MinioManager.PutObject(
		ctx,
		config.GlobalServerConfig.Minio.BucketName,
		fmt.Sprintf(consts.MinioObjectName, roomId, dir[len(dir)-1]),
		f, fileInfo.Size(),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		klog.Error("Record: Failed to put object: ", err)
		return
	}
	klog.Info("Saved file: ", filepath)
	// 删除临时文件。
	err = os.Remove(filepath)
	if err != nil {
		klog.Error("Record: Failed to remove file: ", err)
		return
	}
	// 更新数据库，标记为已录制。

	err = s.IsRecorded(ctx, roomId)
	if err != nil {
		klog.Error("Record: Failed to update lecture IsRecorded: ", err)
		return
	}
}

// GetHistoryLecture implements the LectureServiceImpl interface.
// 获取历史课程，通过房间号获取视频和音频文件，然后播放。
func (s *LectureServiceImpl) GetHistoryLecture(ctx context.Context, request *lecture.GetHistoryLectureRequest) (*lecture.GetHistoryLectureResponse, error) {
	if ok, err := s.CheckRecord(ctx, request.RoomId); !ok {
		klog.Error("GetHistoryLecture: Failed to check record: ", err)
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorRecordNotFound(err.Error()),
		}, nil
	}
	pc, err := s.WebrtcAPI.NewPeerConnection(*s.peerConnectionConfig)
	if err != nil {
		klog.Error("GetHistoryLecture: Failed to create PeerConnection: ", err)
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorCreatePeerConnection(err.Error()),
		}, nil
	}

	// 获取视频流
	ivf, header, err := s.GetIVFStream(ctx, request.RoomId)
	if err != nil {

		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorGetIVFStream(err.Error()),
		}, nil
	}

	// 获取音频流
	ogg, _, err := s.GetOGGStream(ctx, request.RoomId)
	if err != nil {
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorGetOGGStream(err.Error()),
		}, nil
	}

	// 创建视频轨道
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		panic(err)
	}
	videoRtpSender, err := pc.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	// 启动协程读取 RTCP 包（为了处理 NACK/PLI 等）
	s.goroutinePool.Submit(func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := videoRtpSender.Read(rtcpBuf); err != nil {
				return
			}
		}
	})

	// 音频轨道
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		panic(err)
	}
	audioRtpSender, err := pc.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// 启动协程读取 RTCP 包
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := audioRtpSender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()
	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	// 处理视频帧
	s.goroutinePool.Submit(func() {

		// 等待 ICE 连接
		<-iceConnectedCtx.Done()

		ticker := time.NewTicker(
			time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1400),
		)
		defer ticker.Stop()

		for ; true; <-ticker.C {
			frame, _, err := ivf.ParseNextFrame()
			if errors.Is(err, io.EOF) {
				klog.Info("视频流结束")
				return
			}
			if err != nil {
				klog.Error("GetHistoryLecture: Failed to parse IVF frame: ", err)
				return
			}
			if err := videoTrack.WriteSample(media.Sample{Data: frame, Duration: time.Second}); err != nil {
				klog.Error("GetHistoryLecture: Failed to write video sample: ", err)
				return
			}
		}
	})

	// 处理音频页
	s.goroutinePool.Submit(func() {
		var lastGranule uint64
		<-iceConnectedCtx.Done()
		ticker := time.NewTicker(time.Millisecond * 20)
		defer ticker.Stop()

		for ; true; <-ticker.C {
			pageData, pageHeader, err := ogg.ParseNextPage()
			if errors.Is(err, io.EOF) {
				klog.Info("音频流结束")
				return
			}
			if err != nil {
				klog.Error("GetHistoryLecture: Failed to parse OGG page: ", err)
				return
			}
			sampleCount := float64(pageHeader.GranulePosition - lastGranule)
			lastGranule = pageHeader.GranulePosition
			sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

			if err := audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); err != nil {
				klog.Error("GetHistoryLecture: Failed to write audio sample: ", err)
				return
			}
		}
	})

	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateConnected {
			klog.Info("ICE 连接成功")
			iceConnectedCtxCancel()
		}
	})

	// 监听 PeerConnection 状态
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		klog.Info("PeerConnection 状态变化：", state.String())
		if state == webrtc.PeerConnectionStateFailed {
			klog.Error("PeerConnection 状态变化：连接失败，退出")
			pc.Close()
			return
		}
		if state == webrtc.PeerConnectionStateClosed {
			klog.Error("PeerConnection 状态变化：连接关闭，退出")
			pc.Close()
			return
		}
		if state == webrtc.PeerConnectionStateDisconnected {
			klog.Error("PeerConnection 状态变化：连接断开，退出")
			pc.Close()
			return
		}
	})

	offer := webrtc.SessionDescription{}
	util.Decode(request.Offer, &offer)

	// 设置远端 SDP
	if err = pc.SetRemoteDescription(offer); err != nil {
		klog.Error("GetHistoryLecture: Failed to set remote description: ", err)
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorSetRemoteDescription(err.Error()),
		}, nil
	}

	// 创建并设置本地 SDP Answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		klog.Error("GetHistoryLecture: Failed to create answer: ", err)
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorCreateAnswer(err.Error()),
		}, nil
	}
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(answer); err != nil {
		klog.Error("GetHistoryLecture: Failed to set local description: ", err)
		return &lecture.GetHistoryLectureResponse{
			Response: rsp.ErrorSetLocalDescription(err.Error()),
		}, nil
	}
	<-gatherComplete

	return &lecture.GetHistoryLectureResponse{
		Response: rsp.OK(),
		Answer:   util.Encode(pc.LocalDescription()),
	}, nil
}

func (s *LectureServiceImpl) GetIVFStream(ctx context.Context, roomId int64) (*ivfreader.IVFReader, *ivfreader.IVFFileHeader, error) {
	objectName := fmt.Sprintf(consts.MinioObjectName, roomId, "video.ivf")

	// 从 minio 获取对象
	object, err := s.MinioManager.GetObject(ctx, config.GlobalServerConfig.Minio.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 创建 ivf 读取器
	ivf, header, err := ivfreader.NewWith(object)
	if err != nil {
		object.Close()
		return nil, nil, err
	}

	return ivf, header, nil
}

func (s *LectureServiceImpl) GetOGGStream(ctx context.Context, roomId int64) (*oggreader.OggReader, *oggreader.OggHeader, error) {
	objectName := fmt.Sprintf(consts.MinioObjectName, roomId, "audio.ogg")

	// 从 minio 获取对象
	object, err := s.MinioManager.GetObject(ctx, config.GlobalServerConfig.Minio.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 创建 OGG 读取器
	ogg, header, err := oggreader.NewWith(object)
	if err != nil {
		object.Close()
		return nil, nil, err
	}

	return ogg, header, nil
}

// RandomSelect implements the LectureServiceImpl interface.
// 随机点名指定数量的人
func (s *LectureServiceImpl) RandomSelect(ctx context.Context, request *lecture.RandomSelectRequest) (resp *lecture.RandomSelectResponse, err error) {
	value, ok := s.Sessions.Load(request.RoomId)
	if !ok {
		return &lecture.RandomSelectResponse{
			Response: rsp.ErrorRoomNotExists("Room Not Found"),
		}, nil
	}
	session := value.(*LectureSession)
	if session.HostId != request.UserId {
		return &lecture.RandomSelectResponse{
			Response: rsp.ErrorNotTheOwner(),
		}, nil
	}
	var keys []int64

	session.Audiences.Range(func(key, _ any) bool {
		keys = append(keys, key.(int64))
		return true
	})

	if len(keys) == 0 {
		return &lecture.RandomSelectResponse{
			Response: rsp.ErrorNoAudience(),
		}, nil
	}

	// 参数校验
	n := int(request.Number)
	if n > len(keys) {
		n = len(keys)
	} else if n <= 0 {
		n = 1
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Shuffle 可以用来打乱切片，来得到随机的结果
	r.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	// 返回前n个
	selected := keys[:n]

	return &lecture.RandomSelectResponse{
		Response:    rsp.OK(),
		SelectedIds: selected,
	}, nil
}
