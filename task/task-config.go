package task

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"one-api/common"
)

func InitCron() {
	c := cron.New(cron.WithSeconds())

	// 添加定时任务
	_, err := c.AddFunc("0 0 * * * *", func() {
		RefreshAccessToken()
	})
	if err != nil {
		common.SysError("定时任务初始化失败")
	}
	c.Start()
	common.SysLog(fmt.Sprintf("定时任务初始化完成"))
}
