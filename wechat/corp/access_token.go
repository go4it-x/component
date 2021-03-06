package corp

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"log"
	"time"
)

type Param struct {
	AccessTokenPath string `json:"accessTokenPath"`
	AgentId         string `json:"agentId"`
	CorpId          string `json:"corpId"`
	CorpSecret      string `json:"corpSecret"`
}

const tokenUrl = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%v&corpsecret=%v"

// accessToken 获取token
func accessToken(param Param) (accessToken string, err error) {
	if param.AccessTokenPath == "" {
		param.AccessTokenPath = "."
	}

	viper.SetConfigName("access_token")
	viper.SetConfigType("json")
	viper.AddConfigPath(param.AccessTokenPath)
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err = viper.SafeWriteConfig(); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println(err)
			return
		}
	}

	dueTime := viper.GetInt64("expires_in")
	if dueTime > time.Now().Unix() {
		accessToken = viper.GetString("access_token")
		return
	}

	// ============ access_token 已过期，重新获取============
	type Result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	var res Result
	_, err = resty.New().R().
		SetResult(&res).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf(tokenUrl, param.CorpId, param.CorpSecret))
	if err != nil {
		return
	}

	if res.ErrCode != 0 {
		log.Println(res)
		return "", errors.New(res.ErrMsg)
	}

	now := time.Now().Unix()
	accessToken = res.AccessToken
	viper.Set("access_token", accessToken)
	viper.Set("expires_in", now+7150)
	viper.Set("updated_at", now)
	err = viper.WriteConfig()

	return
}
