const WebSocket = require('ws');
const serverConfig = require("./serverConfig")
const constants = require("./constants")

class WsManager {
  constructor(props) {
    this.boundRoomId = null;
    this.intAuthToken = props.intAuthToken;
  }

  sendSafely(msgStr) {
    const instance = this;
    /**
    * - "If the data can't be sent (for example, because it needs to be buffered but the buffer is full), the socket is closed automatically."
    *
    * from https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/send.
    */
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
}

module.exports = WsManager;

function _base64ToUint8Array(base64) {
  var binary_string = instance.atob(base64);
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

