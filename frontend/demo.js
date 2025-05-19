/* eslint-env browser */

// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

const log = msg => {
  document.getElementById('logs').innerHTML += msg + '<br>'
}

async function createSession(isPublisher) {
  const pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: 'stun:stun.l.google.com:19302'
      }
    ]
  })
  pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
  pc.onicecandidate = event => {
    if (event.candidate === null) {
      document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
    }
  }

if (isPublisher) {
  // 添加屏幕共享按钮
  const screenShareBtn = document.createElement('button');
  screenShareBtn.textContent = '共享屏幕';
  screenShareBtn.onclick = async () => {
    try {
      // 获取屏幕共享流
      const screenStream = await navigator.mediaDevices.getDisplayMedia({ 
        video: true,
        audio: true 
      });
      
      // 停止之前的所有轨道
      const senders = pc.getSenders();
      senders.forEach((sender) => {
        if(sender.track) {
          sender.track.stop();
        }
      });
      
      // 清空之前的轨道
      pc.removeTrack(senders[0]);
      pc.removeTrack(senders[1]);
      
      // 添加新的屏幕共享轨道
      screenStream.getTracks().forEach(track => {
        pc.addTrack(track, screenStream);
      });
      
      // 本地预览
      document.getElementById('video1').srcObject = screenStream;
      
      // 当用户停止共享屏幕时切换回摄像头
      screenStream.getVideoTracks()[0].onended = async () => {
        const cameraStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
        const senders = pc.getSenders();
        senders.forEach((sender) => {
          if(sender.track) {
            sender.track.stop();
          }
        });
        pc.removeTrack(senders[0]);
        pc.removeTrack(senders[1]);
        cameraStream.getTracks().forEach(track => pc.addTrack(track, cameraStream));
        document.getElementById('video1').srcObject = cameraStream;
      };
      
    } catch (err) {
      log('获取屏幕共享失败: ' + err);
    }
  };
  document.body.appendChild(screenShareBtn);
  
  // 默认使用摄像头和麦克风
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
      stream.getTracks().forEach(track => pc.addTrack(track, stream));
      document.getElementById('video1').srcObject = stream;
      pc.createOffer()
        .then(desc => pc.setLocalDescription(desc))
        .catch(log);
    })
    .catch(log);
} else {
  // 订阅端：主动创建接收 transceiver，指定 video 和 audio
  pc.addTransceiver('video', { direction: 'recvonly' });
  pc.addTransceiver('audio', { direction: 'recvonly' });

  pc.createOffer()
    .then(desc => pc.setLocalDescription(desc))
    .catch(log);

  // 收到远端 Track 时，将流绑定到 <video> 标签（自带音频播放）
  pc.ontrack = function (event) {
    const el = document.getElementById('video1');
    // event.streams[0] 中同时包含了视频和音频轨道
    el.srcObject = event.streams[0];
    el.autoplay = true;
    el.controls = true;
  };
}


  window.startSession = () => {
    const sd = document.getElementById('remoteSessionDescription').value
    if (sd === '') {
      return alert('Session Description must not be empty')
    }

    try {
      pc.setRemoteDescription(JSON.parse(atob(sd)))
    } catch (e) {
      alert(e)
    }
  }

  window.copySDP = () => {
    const browserSDP = document.getElementById('localSessionDescription')

    browserSDP.focus()
    browserSDP.select()

    try {
      const successful = document.execCommand('copy')
      const msg = successful ? 'successful' : 'unsuccessful'
      log('Copying SDP was ' + msg)
    } catch (err) {
      log('Unable to copy SDP ' + err)
    }
  }

  const btns = document.getElementsByClassName('createSessionButton')
  for (let i = 0; i < btns.length; i++) {
    btns[i].style = 'display: none'
  }

  document.getElementById('signalingContainer').style = 'display: block'
}
