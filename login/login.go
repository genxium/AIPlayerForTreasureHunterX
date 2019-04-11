package login

import (
	C "AI/constants"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type RespGetCaptcha struct {
	Ret                        int    `json:"ret"`
	SmsLoginCaptcha            string `json:"smsLoginCaptcha"`
	GetSmsCaptchaRespErrorCode int    `json:"getSmsCaptchaRespErrorCode"`
}

type RespSmsLogin struct {
	Ret         int    `json:"ret"`
	Token       string `json:"intAuthToken"`
	ExpiresAt   int64  `json:"expiresAt"`
	PlayerID    int    `json:"playerId"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

func GetCaptchaByName(botName string) RespGetCaptcha {
	var respGetCaptcha RespGetCaptcha
	{
    pathGetCaptcha := C.SERVER.PROTOCOL + "://" + C.SERVER.HOST + C.SERVER.PORT + C.API + C.PLAYER + C.VERSION + C.SMS_CAPTCHA + C.GET
		fmt.Println(pathGetCaptcha)
		resp, err := http.Get(pathGetCaptcha + "?phoneNum=" + botName + "&phoneCountryCode=86")
		if err != nil {
			// handle error
			fmt.Println("GetCaptchaByName error!")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("body: %s \n", body)
		json.Unmarshal(body, &respGetCaptcha)
	}
	fmt.Println("1111111")
	fmt.Println(respGetCaptcha)
	return respGetCaptcha
}

func GetIntAuthTokenByCaptcha(botName string, captcha string) (token string, playerId int) {
	var respSmsLogin RespSmsLogin
	{
    pathSmsLogin := C.SERVER.PROTOCOL + "://" + C.SERVER.HOST + C.SERVER.PORT + C.API + C.PLAYER + C.VERSION + C.SMS_CAPTCHA + C.LOGIN
		fmt.Println(pathSmsLogin)
		resp, err := http.PostForm(pathSmsLogin, url.Values{"smsLoginCaptcha": {captcha}, "phoneNum": {botName}, "phoneCountryCode": {"86"}})
		if err != nil {
			// handle error
			fmt.Println("GetIntAuthTokenByCaptcha error!")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &respSmsLogin)
	}
	fmt.Printf("RESPSMSLOGIN: %v \n", respSmsLogin)
	return respSmsLogin.Token, respSmsLogin.PlayerID
}

func GetIntAuthTokenByBotName(botName string) (token string, playerId int) {
	captcha := GetCaptchaByName(botName).SmsLoginCaptcha
	intAuthToken, playerId := GetIntAuthTokenByCaptcha(botName, captcha)
	return intAuthToken, playerId
}
