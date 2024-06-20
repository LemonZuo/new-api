package task

import (
	"fmt"
	"one-api/common"
	"one-api/model"
	"one-api/relay/channel/openai"
)

func RefreshAccessToken() {
	channels, err := model.GetOpenAIAccessTokenWillExpireChannel()
	if err != nil {
		// 查询数据失败
		common.SysError(fmt.Sprintf("查询待更新的渠道数据失败"))
		return
	}
	if len(channels) == 0 {
		common.SysError(fmt.Sprintf("待更新的渠道数据为空"))
		return
	}
	for _, channel := range channels {
		common.SysLog(fmt.Sprintf("开始自动刷新OPENAI AK, channelId: %d, RT: %s", channel.Id, channel.OpenAIRefreshToken))
		res, err := openai.RefreshAccessToken(channel.OpenAIRefreshToken)
		if err != nil {
			common.SysError(fmt.Sprintf("自动刷新OPENAI AK失败, channelId: %d, error: %s", channel.Id, err.Error()))
			continue
		}
		channel.Key = res.AccessToken
		channel.OpenAIAccessTokenExpiresTime = common.GetTimestamp() + res.ExpiresIn
		err = channel.Update()
		if err != nil {
			common.SysError(fmt.Sprintf("自动刷新OPENAI AK,更新数据库失败, channelId: %d, error: %s", channel.Id, err.Error()))
			continue
		}
		common.SysLog(fmt.Sprintf("自动刷新OPENAI AK 成功, channelId: %d", channel.Id))
	}
}
