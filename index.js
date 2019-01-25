const httpRequest = require("./httpRequest");
const wsManager = require("./wsManager");

function gameStart() {
  const httpRequestInst = new httpRequest();
  const getSmsCaptcharParams = {
    phoneNum: "add",
    phoneCountryCode: "86",
  };
  httpRequestInst.get("/SmsCaptcha/get?phoneNum=add&phoneCountryCode=86", getSmsCaptcharParams)
  .then(resp => {
    const smsCaptchaLoginParams = Object.assign(getSmsCaptcharParams, {smsLoginCaptcha: resp.smsLoginCaptcha})
    console.log("smsCaptchaLoginParams:" + JSON.stringify(smsCaptchaLoginParams))
    return httpRequestInst.postForm("/SmsCaptcha/login", smsCaptchaLoginParams)
    .then(resp => {
      const wsManagerIns = new wsManager(resp)
      wsManagerIns.initPersistentSessionClient(() => {
        console.log("connect success!")
      })
    })
  })
  .catch(e => {
    console.log(e)
  })
}

gameStart()

