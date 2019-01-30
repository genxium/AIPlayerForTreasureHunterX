const WebSocket = require('ws');
const serverConfig = require("./serverConfig")
const constants = require("./constants")
const protobuf = require('protobufjs');
const atob = require('atob');


class WsManager {
  constructor(props) {
    const instance = this;
    this.boundRoomId = null;
    this.time = 0;
    this.selfPlayerInfo = Object.assign(props, {x:297.57500000000005, y:799.395});
    this.recentFrameCacheCurrentSize = 0;
    this.recentFrameCacheMaxCount = 2048;
    this.recentFrameCache = {};
    this.clientUpsyncFps = 20;
    this.intAuthToken = props.intAuthToken;
    this.lastRoomDownsyncFrameId = 0;
    this.ALL_BATTLE_STATES = {
      WAITING: 0,
      IN_BATTLE: 1,
      IN_SETTLEMENT: 2,
      IN_DISMISSAL: 3,
    };
    this.battleState = this.ALL_BATTLE_STATES.WAITING;
    protobuf.load("./room_downsync_frame.proto", function(err, root) {
      if (err) {
        console.log(err)
      }
      instance.RoomDownsyncFrame = root.lookupType("models.RoomDownsyncFrame");
    })
  }

  sendSafely(msgStr) {
    const instance = this;
    if (null == instance.clientSession || instance.clientSession.readyState != WebSocket.OPEN) return false;
    instance.clientSession.send(msgStr);
  }

  closeWSConnection () {
    const instance = this;
    if (null == instance.clientSession || instance.clientSession.readyState != WebSocket.OPEN) return;
    console.log(`Closing "instance.clientSession" from the client-side.`);
    instance.clientSession.close();
  }

  handleHbRequirements(resp) {
    console.log(resp)
    const instance = this;
    if (constants.RET_CODE.OK != resp.ret) return;
    if (null == instance.boundRoomId) {
      instance.boundRoomId = resp.data.boundRoomId;
      //TODO
      //expiresAt = Date.now() + 10 * 60 * 1000; //TODO: hardcoded, boundRoomId过期时间
    }
  };

  clientSessionPingInterval() {
    const instance = this;
    setInterval(() => {
      if (clientSession.readyState != WebSocket.OPEN) return;
      const param = {
        msgId: Date.now(),
        act: "HeartbeatPing",
        data: {
          clientTimestamp: Date.now()
        }
      };
      instance.sendSafely(JSON.stringify(param));
    }, resp.data.intervalToPing);
  }
 
  handleHbPong (resp) {
    if (constants.RET_CODE.OK != resp.ret) return;
    // TBD.
  };

  initPersistentSessionClient (onopenCb) {
    const instance = this;
    if (instance.clientSession && instance.clientSession.readyState == WebSocket.OPEN) {
      if (null != onopenCb) {
        onopenCb();
      }
      return;
    }

    let urlToConnect = serverConfig.PROTOCOL.replace('http', 'ws') + '://' + serverConfig.HOST + ":" + serverConfig.PORT + serverConfig.WS_PATH_PREFIX + "?intAuthToken=" + instance.intAuthToken;

    let expectedRoomId = instance.boundRoomId;
    if (expectedRoomId) {
      console.log("initPersistentSessionClient with expectedRoomId == " +  expectedRoomId);
      urlToConnect = urlToConnect + "&expectedRoomId=" + expectedRoomId;
    } else {
      console.error("constructor valuation expectedRoomId fail")
    }

    const clientSession = new WebSocket(urlToConnect);

    clientSession.onopen = function(event) {
      console.log("The WS clientSession is opened.");
      instance.clientSession = clientSession;
      if (null == onopenCb) return;
      //TODO
      onopenCb();
    };

    clientSession.onmessage = function(event) {
      const resp = JSON.parse(event.data)
      switch (resp.act) {
        case "HeartbeatRequirements":
          instance.handleHbRequirements(resp);
          break;
        case "HeartbeatPong":
          instance.handleHbPong(resp);
          break;
        case "RoomDownsyncFrame":
          //TODO
          if (instance.handleRoomDownsyncFrame) {
            const typedArray = _base64ToUint8Array(resp.data);
            const parsedRoomDownsyncFrame = instance.RoomDownsyncFrame.decode(typedArray);
            console.log("parsedRoomDownsyncFrame" + JSON.stringify(parsedRoomDownsyncFrame))
            instance.handleRoomDownsyncFrame(parsedRoomDownsyncFrame);
          }
          break;
        case "Ready": {
          //TODO
          if (instance.handleGameReadyResp) {
            instance.handleGameReadyResp(resp);
          }
          break;
        }
        default:
          console.log(`${JSON.stringify(resp)}`);
          break;
      }
    };

    clientSession.onerror = function(event) {
      console.error(`Error caught on the WS clientSession:`, event);
      if (instance.clientSessionPingInterval) {
        //TODO clearInterval
        clearInterval(instance.clientSessionPingInterval);
      }
      //TODO handleClientSessionCloseOrError
      if (instance.handleClientSessionCloseOrError) {
        instance.handleClientSessionCloseOrError();
      }
    };

    clientSession.onclose = function(event) {
      console.log(`The WS clientSession is closed:`, event);
      if (instance.clientSessionPingInterval) {
        clearInterval(instance.clientSessionPingInterval);
      }
      if (false == event.wasClean) {
        // Chrome doesn't allow the use of "CustomCloseCode"s (yet) and will callback with a "WebsocketStdCloseCode 1006" and "false == event.wasClean" here. See https://tools.ietf.org/html/rfc6455#section-7.4 for more information.
        if (instance.handleClientSessionCloseOrError) {
          instance.handleClientSessionCloseOrError();
        }
      } else {
        switch (event.code) {
          case constants.RET_CODE.PLAYER_NOT_FOUND:
          case constants.RET_CODE.PLAYER_CHEATING:
            instance.boundRoomId = null;
            break;
          default:
            break;
        }

        if (instance.handleClientSessionCloseOrError) {
          instance.handleClientSessionCloseOrError();
        }
      }
    };
  };

  handleRoomDownsyncFrame(diffFrame) {
    const self = this;
    const ALL_BATTLE_STATES = self.ALL_BATTLE_STATES;
    if (ALL_BATTLE_STATES.WAITING != self.battleState && ALL_BATTLE_STATES.IN_BATTLE != self.battleState && ALL_BATTLE_STATES.IN_SETTLEMENT != self.battleState) return;
    const refFrameId = diffFrame.refFrameId;
    const frameId = diffFrame.id;
    if (frameId <= self.lastRoomDownsyncFrameId) {
      return;
    }
    const isInitiatingFrame = (0 > self.recentFrameCacheCurrentSize || 0 == refFrameId);
    const cachedFullFrame = self.recentFrameCache[refFrameId];
    if (
      !isInitiatingFrame
    ) {
      return;
    }

    if (isInitiatingFrame && 0 == refFrameId) {
      self._onResyncCompleted();
    }
    let countdownNanos = diffFrame.countdownNanos;
    if (countdownNanos < 0)
      countdownNanos = 0;
    const countdownSeconds = parseInt(countdownNanos / 1000000000);
    if (isNaN(countdownSeconds)) {
      cc.log(`countdownSeconds is NaN for countdownNanos == ${countdownNanos}.`);
    }
    const roomDownsyncFrame = (
    (isInitiatingFrame)
      ?
      diffFrame
      :
      self._generateNewFullFrame(cachedFullFrame, diffFrame)
    );
    if (countdownNanos <= 0) {
      self.onBattleStopped(roomDownsyncFrame.players);
      return;
    }
    self._dumpToFullFrameCache(roomDownsyncFrame);
    const sentAt = roomDownsyncFrame.sentAt;


    //update players Info
    const players = roomDownsyncFrame.players;
    const playerIdStrList = Object.keys(players);
    self.otherPlayerCachedDataDict = {};
    for (let i = 0; i < playerIdStrList.length; ++i) {
      const k = playerIdStrList[i];
      const playerId = parseInt(k);
      if (playerId == self.selfPlayerInfo.id) {
        const immediateSelfPlayerInfo = players[k];
        console.log("immediateSelfPlayerInfo:" + immediateSelfPlayerInfo)
        Object.assign(self.selfPlayerInfo, {
          x: immediateSelfPlayerInfo.x,
          y: immediateSelfPlayerInfo.y,
          speed: immediateSelfPlayerInfo.speed,
          battleState: immediateSelfPlayerInfo.battleState,
          score: immediateSelfPlayerInfo.score,
          joinIndex: immediateSelfPlayerInfo.joinIndex,
        });
        continue;
      }
      const anotherPlayer = players[k];
      self.otherPlayerCachedDataDict[playerId] = anotherPlayer;
    }

    //update pumpkin Info 
    self.pumpkinInfoDict = {};
    const pumpkin = roomDownsyncFrame.pumpkin;
    const pumpkinsLocalIdStrList = Object.keys(pumpkin);
    for (let i = 0; i < pumpkinsLocalIdStrList.length; ++i) {
      const k = pumpkinsLocalIdStrList[i];
      const pumpkinLocalIdInBattle = parseInt(k);
      const pumpkinInfo = pumpkin[k];
      self.pumpkinInfoDict[pumpkinLocalIdInBattle] = pumpkinInfo;
    }
    

    //update treasureInfoDict
    self.treasureInfoDict = {};
    const treasures = roomDownsyncFrame.treasures;
    const treasuresLocalIdStrList = Object.keys(treasures);
    for (let i = 0; i < treasuresLocalIdStrList.length; ++i) {
      const k = treasuresLocalIdStrList[i];
      const treasureLocalIdInBattle = parseInt(k);
      const treasureInfo = treasures[k];
      self.treasureInfoDict[treasureLocalIdInBattle] = treasureInfo;
    }

    //update acceleratorInfoDict
    self.acceleratorInfoDict = {};
    const accelartors = roomDownsyncFrame.speedShoes;
    const accLocalIdStrList = Object.keys(accelartors);
    for (let i = 0; i < accLocalIdStrList.length; ++i) {
      const k = accLocalIdStrList[i];
      const accLocalIdInBattle = parseInt(k);
      const accInfo = accelartors[k];
      self.acceleratorInfoDict[accLocalIdInBattle] = accInfo;
    }

    //update trapInfoDict
    self.trapInfoDict = {};
    const traps = roomDownsyncFrame.traps;
    const trapsLocalIdStrList = Object.keys(traps);
    for (let i = 0; i < trapsLocalIdStrList.length; ++i) {
      const k = trapsLocalIdStrList[i];
      const trapLocalIdInBattle = parseInt(k);
      const trapInfo = traps[k];
      self.trapInfoDict[trapLocalIdInBattle] = trapInfo;
    }

    self.trapBulletInfoDict = {};
    const bullets = roomDownsyncFrame.bullets;
    const bulletsLocalIdStrList = Object.keys(bullets);
    for (let i = 0; i < bulletsLocalIdStrList.length; ++i) {
      const k = bulletsLocalIdStrList[i];
      const bulletLocalIdInBattle = parseInt(k);
      const bulletInfo = bullets[k];
      self.trapBulletInfoDict[bulletLocalIdInBattle] = bulletInfo;
    }

    if (0 == self.lastRoomDownsyncFrameId) {
      self.battleState = ALL_BATTLE_STATES.IN_BATTLE;
      if (1 == frameId) {
        console.log("game start")
      }
      self.onBattleStarted();
    }
    self.lastRoomDownsyncFrameId = frameId;
  };
  transitToState(s) {
    const self = this;
    self.state = s;
  }
  
  _dumpToFullFrameCache (fullFrame) {
    const self = this;
    while (self.recentFrameCacheCurrentSize >= self.recentFrameCacheMaxCount) {
      // Trick here: never evict the "Zero-th Frame" for resyncing!
      const toDelFrameId = Object.keys(self.recentFrameCache)[1];
      // cc.log("toDelFrameId is " + toDelFrameId + ".");
      delete self.recentFrameCache[toDelFrameId];
      --self.recentFrameCacheCurrentSize;
    }
    self.recentFrameCache[fullFrame.id] = fullFrame;
    ++self.recentFrameCacheCurrentSize;
  }

  _onResyncCompleted() {
    if (false == this.resyncing) return;
    console.log(`_onResyncCompleted`);
    this.resyncing = false;
    if (null != this.resyncingHintPopup && this.resyncingHintPopup.parent) {
      this.resyncingHintPopup.parent.removeChild(this.resyncingHintPopup);
    }
  }

  _generateNewFullFrame(refFullFrame, diffFrame) {
    let newFullFrame = {
      id: diffFrame.id,
      treasures: refFullFrame.treasures,
      traps: refFullFrame.traps,
      bullets: refFullFrame.bullets,
      players: refFullFrame.players,
      speedShoes: refFullFrame.speedShoes,
      pumpkin: refFullFrame.pumpkin,
    };
    const players = diffFrame.players;
    const playersLocalIdStrList = Object.keys(players);
    for (let i = 0; i < playersLocalIdStrList.length; ++i) {
      const k = playersLocalIdStrList[i];
      const playerId = parseInt(k);
      if (true == diffFrame.players[playerId].removed) {
        // cc.log(`Player id == ${playerId} is removed.`);
        delete newFullFrame.players[playerId];
      } else {
        newFullFrame.players[playerId] = diffFrame.players[playerId];
      }
    }

    const pumpkin = diffFrame.pumpkin;
    const pumpkinsLocalIdStrList = Object.keys(pumpkin);
    for (let i = 0; i < pumpkinsLocalIdStrList.length; ++i) {
      const k = pumpkinsLocalIdStrList[i];
      const pumpkinLocalIdInBattle = parseInt(k);
      if (true == diffFrame.pumpkin[pumpkinLocalIdInBattle].removed) {
        delete newFullFrame.pumpkin[pumpkinLocalIdInBattle];
      } else {
        newFullFrame.pumpkin[pumpkinLocalIdInBattle] = diffFrame.pumpkin[pumpkinLocalIdInBattle];
      }
    }

    const treasures = diffFrame.treasures;
    const treasuresLocalIdStrList = Object.keys(treasures);
    for (let i = 0; i < treasuresLocalIdStrList.length; ++i) {
      const k = treasuresLocalIdStrList[i];
      const treasureLocalIdInBattle = parseInt(k);
      if (true == diffFrame.treasures[treasureLocalIdInBattle].removed) {
        // cc.log(`Treasure with localIdInBattle == ${treasureLocalIdInBattle} is removed.`);
        delete newFullFrame.treasures[treasureLocalIdInBattle];
      } else {
        newFullFrame.treasures[treasureLocalIdInBattle] = diffFrame.treasures[treasureLocalIdInBattle];
      }
    }

    const speedShoes = diffFrame.speedShoes;
    const speedShoesLocalIdStrList = Object.keys(speedShoes);
    for (let i = 0; i < speedShoesLocalIdStrList.length; ++i) {
      const k = speedShoesLocalIdStrList[i];
      const speedShoesLocalIdInBattle = parseInt(k);
      if (true == diffFrame.speedShoes[speedShoesLocalIdInBattle].removed) {
        // cc.log(`Treasure with localIdInBattle == ${treasureLocalIdInBattle} is removed.`);
        delete newFullFrame.speedShoes[speedShoesLocalIdInBattle];
      } else {
        newFullFrame.speedShoes[speedShoesLocalIdInBattle] = diffFrame.speedShoes[speedShoesLocalIdInBattle];
      }
    }

    const traps = diffFrame.traps;
    const trapsLocalIdStrList = Object.keys(traps);
    for (let i = 0; i < trapsLocalIdStrList.length; ++i) {
      const k = trapsLocalIdStrList[i];
      const trapLocalIdInBattle = parseInt(k);
      if (true == diffFrame.traps[trapLocalIdInBattle].removed) {
        // cc.log(`Trap with localIdInBattle == ${trapLocalIdInBattle} is removed.`);
        delete newFullFrame.traps[trapLocalIdInBattle];
      } else {
        newFullFrame.traps[trapLocalIdInBattle] = diffFrame.traps[trapLocalIdInBattle];
      }
    }

    const bullets = diffFrame.bullets;
    const bulletsLocalIdStrList = Object.keys(bullets);
    for (let i = 0; i < bulletsLocalIdStrList.length; ++i) {
      const k = bulletsLocalIdStrList[i];
      const bulletLocalIdInBattle = parseInt(k);
      if (true == diffFrame.bullets[bulletLocalIdInBattle].removed) {
        cc.log(`Bullet with localIdInBattle == ${bulletLocalIdInBattle} is removed.`);
        delete newFullFrame.bullets[bulletLocalIdInBattle];
      } else {
        newFullFrame.bullets[bulletLocalIdInBattle] = diffFrame.bullets[bulletLocalIdInBattle];
      }
    }

    const accs = diffFrame.speedShoes;
    const accsLocalIdStrList = Object.keys(accs);
    for (let i = 0; i < accsLocalIdStrList.length; ++i) {
      const k = accsLocalIdStrList[i];
      const accLocalIdInBattle = parseInt(k);
      if (true == diffFrame.speedShoes[accLocalIdInBattle].removed) {
        delete newFullFrame.speedShoes[accLocalIdInBattle];
      } else {
        newFullFrame.speedShoes[accLocalIdInBattle] = diffFrame.speedShoes[accLocalIdInBattle];
      }
    }
    return newFullFrame;
  }
  onBattleStopped(players) {
    const self = this;
    self.battleState = ALL_BATTLE_STATES.IN_SETTLEMENT;
  }
  onBattleStarted() {
    const self = this;
    self.upsyncLoopInterval = setInterval(self._onPerUpsyncFrame.bind(self), self.clientUpsyncFps);
  }
  _onPerUpsyncFrame() {
    const instance = this;
    if (instance.resyncing) return;
    if (null == instance.selfPlayerInfo) return;
    let x = parseFloat(instance.selfPlayerInfo.x + instance.time * 0.3);
    let y = parseFloat(instance.selfPlayerInfo.y);
    const upsyncFrameData = {
      id: instance.selfPlayerInfo.playerId,
      /**
      * WARNING
      *
      * Deliberately NOT upsyncing the `instance.selfPlayerScriptIns.activeDirection` here, because it'll be deduced by other players from the position differences of `RoomDownsyncFrame`s.
      */
      dir: {
        dx: 1,
        dy: 0,
      },
      x: x,
      y: y,
      ackingFrameId: instance.lastRoomDownsyncFrameId,
    };
    const wrapped = {
      msgId: Date.now(),
      act: "PlayerUpsyncCmd",
      data: upsyncFrameData,
    }
    instance.sendSafely(JSON.stringify(wrapped));
    instance.time++
  }
}

module.exports = WsManager;

function _base64ToUint8Array(base64) {
  var binary_string = atob(base64);
  var len = binary_string.length;
  var bytes = new Uint8Array(len);
  for (var i = 0; i < len; i++) {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

function _base64ToArrayBuffer(base64) {
  return _base64ToUint8Array(base64).buffer;
}

