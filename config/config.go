package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// GetConfig returns the global config
func GetConfig() *viper.Viper {
	c := viper.New()
	c.SetConfigType("yaml")
	c.SetConfigName("v2board")
	c.AddConfigPath(".")
	c.AutomaticEnv()

	c.SetDefault("debug", true)
	c.SetDefault("appName", "v2board机器人")
	c.SetDefault("traffic", 1024)
	c.SetDefault("isAutoDeleteMsg", true)
	c.SetDefault("telegram.admins", []interface{}{})

	c.SetDefault("redis.host", "localhost:6379")
	c.SetDefault("redis.db", 0)
	c.SetDefault("redis.password", "")
	c.SetDefault("redis.cacheTime", 12)
	c.SetDefault("telegram.key", "")

	replacer := strings.NewReplacer(".", "_")
	c.SetEnvKeyReplacer(replacer)

	if err := c.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	return c
}
func GetKeywordsConfig()  map[string]string {
	viper.SetConfigFile("autoReply.yaml")
    err := viper.ReadInConfig()
    if err != nil {
        return nil
    }

    // 从配置文件中读取关键字和回复内容

    keywords := viper.GetStringMapString("keywords")
	return keywords
}
