package config

import (
	"github.com/BurntSushi/toml"
)

// Config 對應 conf.toml 的結構
type Config struct {
	WebHook string `toml:"web_hook"`
}

// LoadConfig 讀取並解析 conf.toml
func LoadConfig(path string) (*Config, error) {
	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// 建議在使用時，可以搭配這樣呼叫：
// conf, err := config.LoadConfig("conf.toml")
// if err != nil {
//     log.Fatalf("載入設定失敗: %v", err)
// }
// fmt.Println(conf.WebHook)
