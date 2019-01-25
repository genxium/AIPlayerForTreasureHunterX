const request = require("supertest")
const serverConfig = require("./serverConfig")
const constants = require("./constants")

class httpRequest {
  constructor(){
    this.backendAddress = serverConfig.PROTOCOL + '://' + serverConfig.HOST + ":" + serverConfig.PORT + "/api/player/v1";
  }
  postJson(api, params) {
    const instance = this;
    return request(instance.backendAddress)
    .post(api)
    .type('application/json')
    .send(params)
    .then(res => {
      return new Promise((resolve, reject) => {
        if (res && res.body && res.body.ret == constants.RET_CODE.OK) {
          resolve(res.body)  
        } else {
          if(res.body) {
            reject(res.body)
          } else {
            reject(res)
          }
        }
      })
    })
  }
  postForm(api, params) {
    let formPramas = "";
    for(let index in params) {
      formPramas = formPramas + index + "=" + params[index] + "&";
    }
    console.log(formPramas)
    const instance = this;
    return request(instance.backendAddress)
    .post(api)
    .send(formPramas)
    .then(res => {
      return new Promise((resolve, reject) => {
        if (res && res.body && res.body.ret == constants.RET_CODE.OK) {
          resolve(res.body)  
        } else {
          if(res.body) {
            reject(res.body)
          } else {
            reject(res)
          }
        }
      })
    })
  }
  get(api, params) {
    const instance = this;
    return request(instance.backendAddress)
    .get(api)
    .type('application/json')
    .send(params)
    .then(res => {
      return new Promise((resolve, reject) => {
        if (res && res.body && res.body.ret == constants.RET_CODE.OK || res.body.ret == constants.RET_CODE.IS_TEST_ACC) {
          resolve(res.body)  
        } else {
          if(res.body) {
            reject(res.body)
          } else {
            reject(res)
          }
        }
      })
    })
  }
}

module.exports = httpRequest;
