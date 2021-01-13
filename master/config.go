package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//解析json
type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
}

//单例对象 用于配置
var (
	G_config *Config
)

//初始化配置
func InitConfig(fileName string) (err error) {
	var (
		content []byte
		cfg     Config
	)

	//读取配置文件
	if content, err = ioutil.ReadFile(fileName); err != nil {
		fmt.Println(err)
		return
	}

	//json反序列化
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return
	}

	fmt.Printf("port:%d readTimeout:%d writeTimeout:%d\n", cfg.ApiPort, cfg.ApiReadTimeout, cfg.ApiWriteTimeout)

	//赋值单例
	G_config = &cfg

	return err
}
