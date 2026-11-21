package model

// SystemSetting 站点开关，key-value 存储（对齐主仓库 system_settings 的简化版）。
// SSO 的密钥类配置一律走环境变量，本表只存明文开关值，不需要 SecretBox。
type SystemSetting struct {
	Key   string `gorm:"size:64;primaryKey" json:"-"`
	Value string `gorm:"size:255" json:"-"`
}

// TableName 表名。
func (SystemSetting) TableName() string { return "system_settings" }
