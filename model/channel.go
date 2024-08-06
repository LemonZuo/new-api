package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"one-api/common"
	"time"
	"strings"
)

type Channel struct {
	Id                           int     `json:"id"`
	Type                         int     `json:"type" gorm:"default:0"`
	Key                          string  `json:"key" gorm:"not null"`
	OpenAIRefreshToken           string  `json:"openai_refresh_token" gorm:"default:''"`
	OpenAIAccessTokenExpiresTime int64   `json:"openai_access_token_expires_time" gorm:"bigint"`
	OpenAIOrganization           *string `json:"openai_organization"`
	TestModel                    *string `json:"test_model"`
	Status                       int     `json:"status" gorm:"default:1"`
	Name                         string  `json:"name" gorm:"index"`
	Weight                       *uint   `json:"weight" gorm:"default:0"`
	CreatedTime                  int64   `json:"created_time" gorm:"bigint"`
	TestTime                     int64   `json:"test_time" gorm:"bigint"`
	ResponseTime                 int     `json:"response_time"` // in milliseconds
	BaseURL                      *string `json:"base_url" gorm:"column:base_url;default:''"`
	Other                        string  `json:"other"`
	Balance                      float64 `json:"balance"` // in USD
	BalanceUpdatedTime           int64   `json:"balance_updated_time" gorm:"bigint"`
	Models                       string  `json:"models"`
	Group                        string  `json:"group" gorm:"type:varchar(64);default:'default'"`
	UsedQuota                    int64   `json:"used_quota" gorm:"bigint;default:0"`
	ModelMapping                 *string `json:"model_mapping" gorm:"type:varchar(1024);default:''"`
	// MaxInputTokens     *int    `json:"max_input_tokens" gorm:"default:0"`
	StatusCodeMapping *string `json:"status_code_mapping" gorm:"type:varchar(1024);default:''"`
	Priority          *int64  `json:"priority" gorm:"bigint;default:0"`
	AutoBan           *int    `json:"auto_ban" gorm:"default:1"`
	OtherInfo         string  `json:"other_info"`
	Headers           string  `json:"headers" gorm:"type:varchar(1024);default:''"`
	Proxy             string  `json:"proxy" gorm:"type:varchar(1024);default:''"`
}

func (channel *Channel) GetModels() []string {
	if channel.Models == "" {
		return []string{}
	}
	return strings.Split(strings.Trim(channel.Models, ","), ",")
}

func (channel *Channel) GetOtherInfo() map[string]interface{} {
	otherInfo := make(map[string]interface{})
	if channel.OtherInfo != "" {
		err := json.Unmarshal([]byte(channel.OtherInfo), &otherInfo)
		if err != nil {
			common.SysError("failed to unmarshal other info: " + err.Error())
		}
	}
	return otherInfo
}

func (channel *Channel) SetOtherInfo(otherInfo map[string]interface{}) {
	otherInfoBytes, err := json.Marshal(otherInfo)
	if err != nil {
		common.SysError("failed to marshal other info: " + err.Error())
		return
	}
	channel.OtherInfo = string(otherInfoBytes)
}

func (channel *Channel) GetAutoBan() bool {
	if channel.AutoBan == nil {
		return false
	}
	return *channel.AutoBan == 1
}

func (channel *Channel) Save() error {
	return DB.Save(channel).Error
}

func GetAllChannels(startIdx int, num int, selectAll bool, idSort bool) ([]*Channel, error) {
	var channels []*Channel
	var err error
	order := "priority desc"
	if idSort {
		order = "id desc"
	}
	if selectAll {
		err = DB.Order(order).Find(&channels).Error
	} else {
		err = DB.Order(order).Limit(num).Offset(startIdx).Omit("key", "open_ai_refresh_token", "open_ai_access_token_expires_time").Find(&channels).Error
	}
	return channels, err
}

func SearchChannels(keyword string, group string, model string) ([]*Channel, error) {
	var channels []*Channel
	keyCol := "`key`"
	openaiRefreshTokenCol := "`open_ai_refresh_token`"
	openAIAccessTokenExpiresTimeCol := "`open_ai_access_token_expires_time`"
	groupCol := "`group`"
	modelsCol := "`models`"

	// 如果是 PostgreSQL，使用双引号
	if common.UsingPostgreSQL {
		keyCol = `"key"`
		openaiRefreshTokenCol = `"open_ai_refresh_token"`
		openAIAccessTokenExpiresTimeCol = `"open_ai_access_token_expires_time"`
		groupCol = `"group"`
		modelsCol = `"models"`
	}

	// 构造基础查询
	baseQuery := DB.Model(&Channel{}).Omit(keyCol, openaiRefreshTokenCol, openAIAccessTokenExpiresTimeCol)

	// 构造WHERE子句
	var whereClause string
	var args []interface{}
	if group != "" {
		whereClause = "(id = ? OR name LIKE ? OR " + keyCol + " = ?) AND " + groupCol + " = ? AND " + modelsCol + " LIKE ?"
		args = append(args, common.String2Int(keyword), "%"+keyword+"%", keyword, group, "%"+model+"%")
	} else {
		whereClause = "(id = ? OR name LIKE ? OR " + keyCol + " = ?) AND " + modelsCol + " LIKE ?"
		args = append(args, common.String2Int(keyword), "%"+keyword+"%", keyword, "%"+model+"%")
	}

	// 执行查询
	err := baseQuery.Where(whereClause, args...).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func GetChannelById(id int, selectAll bool) (*Channel, error) {
	channel := Channel{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&channel, "id = ?", id).Error
	} else {
		err = DB.Omit("key", "open_ai_refresh_token", "open_ai_access_token_expires_time").First(&channel, "id = ?", id).Error
	}
	return &channel, err
}

func BatchInsertChannels(channels []Channel) error {
	var err error
	err = DB.Create(&channels).Error
	if err != nil {
		return err
	}
	for _, channel_ := range channels {
		err = channel_.AddAbilities()
		if err != nil {
			return err
		}
	}
	return nil
}

func BatchDeleteChannels(ids []int) error {
	// 使用事务 删除channel表和channel_ability表
	tx := DB.Begin()
	err := tx.Where("id in (?)", ids).Delete(&Channel{}).Error
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}
	err = tx.Where("channel_id in (?)", ids).Delete(&Ability{}).Error
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}
	// 提交事务
	tx.Commit()
	return err
}

func (channel *Channel) GetPriority() int64 {
	if channel.Priority == nil {
		return 0
	}
	return *channel.Priority
}

func (channel *Channel) GetWeight() int {
	if channel.Weight == nil {
		return 0
	}
	return int(*channel.Weight)
}

func (channel *Channel) GetBaseURL() string {
	if channel.BaseURL == nil {
		return ""
	}
	return *channel.BaseURL
}

func (channel *Channel) GetModelMapping() string {
	if channel.ModelMapping == nil {
		return ""
	}
	return *channel.ModelMapping
}

func (channel *Channel) GetStatusCodeMapping() string {
	if channel.StatusCodeMapping == nil {
		return ""
	}
	return *channel.StatusCodeMapping
}

func (channel *Channel) Insert() error {
	var err error
	err = DB.Create(channel).Error
	if err != nil {
		return err
	}
	err = channel.AddAbilities()
	return err
}

func (channel *Channel) Update() error {
	var err error
	err = DB.Model(channel).Updates(channel).Error
	if err != nil {
		return err
	}
	DB.Model(channel).First(channel, "id = ?", channel.Id)
	err = channel.UpdateAbilities()
	return err
}

func (channel *Channel) UpdateResponseTime(responseTime int64) {
	err := DB.Model(channel).Select("response_time", "test_time").Updates(Channel{
		TestTime:     common.GetTimestamp(),
		ResponseTime: int(responseTime),
	}).Error
	if err != nil {
		common.SysError("failed to update response time: " + err.Error())
	}
}

func (channel *Channel) UpdateBalance(balance float64) {
	err := DB.Model(channel).Select("balance_updated_time", "balance").Updates(Channel{
		BalanceUpdatedTime: common.GetTimestamp(),
		Balance:            balance,
	}).Error
	if err != nil {
		common.SysError("failed to update balance: " + err.Error())
	}
}

func (channel *Channel) Delete() error {
	var err error
	err = DB.Delete(channel).Error
	if err != nil {
		return err
	}
	err = channel.DeleteAbilities()
	return err
}

func UpdateChannelStatusById(id int, status int, reason string) {
	err := UpdateAbilityStatus(id, status == common.ChannelStatusEnabled)
	if err != nil {
		common.SysError("failed to update ability status: " + err.Error())
	}
	channel, err := GetChannelById(id, true)
	if err != nil {
		// find channel by id error, directly update status
		err = DB.Model(&Channel{}).Where("id = ?", id).Update("status", status).Error
		if err != nil {
			common.SysError("failed to update channel status: " + err.Error())
		}
	} else {
		// find channel by id success, update status and other info
		info := channel.GetOtherInfo()
		info["status_reason"] = reason
		info["status_time"] = common.GetTimestamp()
		channel.SetOtherInfo(info)
		channel.Status = status
		err = channel.Save()
		if err != nil {
			common.SysError("failed to update channel status: " + err.Error())
		}
	}

}

func UpdateChannelUsedQuota(id int, quota int) {
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeChannelUsedQuota, id, quota)
		return
	}
	updateChannelUsedQuota(id, quota)
}

func updateChannelUsedQuota(id int, quota int) {
	err := DB.Model(&Channel{}).Where("id = ?", id).Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
	if err != nil {
		common.SysError("failed to update channel used quota: " + err.Error())
	}
}

func DeleteChannelByStatus(status int64) (int64, error) {
	result := DB.Where("status = ?", status).Delete(&Channel{})
	return result.RowsAffected, result.Error
}

func DeleteDisabledChannel() (int64, error) {
	result := DB.Where("status = ? or status = ?", common.ChannelStatusAutoDisabled, common.ChannelStatusManuallyDisabled).Delete(&Channel{})
	return result.RowsAffected, result.Error
}

func GetOpenAIAccessTokenWillExpireChannel() ([]*Channel, error) {
	var channels []*Channel
	// 计算24小时后的时间戳
	expired := time.Now().Add(24 * time.Hour).Unix()
	// 查询所有在24小时内将会过期,且有的Channel
	err := DB.Where("type = 1 AND open_ai_refresh_token IS NOT NULL AND open_ai_refresh_token != ''  AND open_ai_access_token_expires_time > 0 AND open_ai_access_token_expires_time <= ?", expired).Find(&channels).Error
	return channels, err
}
