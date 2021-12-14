package nacos

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	vinehoo_nacos *VinehooNacosConfig
	ConfigClient  config_client.IConfigClient
)

type NacosClientConfig struct {
	NamespaceId         string `json:"NamespaceId"`
	TimeoutMs           uint64 `json:"TimeoutMs"`
	NotLoadCacheAtStart bool   `json:"NotLoadCacheAtStart"`
	Username            string `json:"Username"`
	Password            string `json:"Password"`
	AccessKey           string `json:"AccessKey"`
	SecretKey           string `json:"SecretKey"`
}

type NacosServerConfig struct {
	IpAddr      string `json:"IpAddr"`
	ContextPath string `json:"ContextPath"`
	Port        uint64 `json:"Port"`
	Scheme      string `json:"Scheme"`
}

type NacosConfig struct {
	ClientConfig NacosClientConfig `json:"ClientConfig"`
	ServerConfig NacosServerConfig `json:"ServerConfig"`
}

type ConfRoot struct {
	Nacos NacosConfig `json:"nacos"`
}

type VinehooNacosConfig struct {
}

type ConfigListenHandler = func(data *string, err error)

func init() {
	var (
		err error
	)
	if vinehoo_nacos, err = NewVinehooNacosConfig("config/config"); err != nil {
		panic(err)
	}
}

// 创建VinehooNacosConfig对象
//
// filename 为nacos配置文件路径
//
// 配置文件格式:
//
// {
// 	"nacos": {
// 		"ClientConfig": {
// 			"NamespaceId": "123434-6569-4a2d-a623-5df9999dd91a",
// 			"TimeoutMs": 5000,
// 			"NotLoadCacheAtStart": true,
// 			"Username": "123",
// 			"Password": "12345678"
// 		},
// 		"ServerConfig": {
// 			"IpAddr": "127.0.0.1",
// 			"ContextPath": "/nacos",
// 			"Port": 8848,
// 			"Scheme": "http"
// 		}
// 	}
// }
func NewVinehooNacosConfig(filename string) (*VinehooNacosConfig, error) {
	vnc := VinehooNacosConfig{}

	f_config, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f_config.Close()

	fileinfo, err := f_config.Stat()
	if err != nil {
		return nil, err
	}

	filesize := fileinfo.Size()
	data_buf := make([]byte, filesize)
	_, err = f_config.Read(data_buf)
	if err != nil {
		return nil, err
	}

	conf_data := ConfRoot{}

	err = json.Unmarshal(data_buf, &conf_data)
	if err != nil {
		return nil, err
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         conf_data.Nacos.ClientConfig.NamespaceId,
		TimeoutMs:           conf_data.Nacos.ClientConfig.TimeoutMs,
		NotLoadCacheAtStart: conf_data.Nacos.ClientConfig.NotLoadCacheAtStart,
	}

	if conf_data.Nacos.ClientConfig.Username == "" {
		clientConfig.AccessKey = conf_data.Nacos.ClientConfig.AccessKey
		clientConfig.SecretKey = conf_data.Nacos.ClientConfig.SecretKey
	} else {
		clientConfig.Username = conf_data.Nacos.ClientConfig.Username
		clientConfig.Password = conf_data.Nacos.ClientConfig.Password
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      conf_data.Nacos.ServerConfig.IpAddr,
			ContextPath: conf_data.Nacos.ServerConfig.ContextPath,
			Port:        conf_data.Nacos.ServerConfig.Port,
			Scheme:      conf_data.Nacos.ServerConfig.Scheme,
		},
	}

	if ConfigClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	); err != nil {
		panic(err)
	}

	return &vnc, nil
}

// 获取配置
func GetString(dataId string, group string, listenHandler ConfigListenHandler) (string, error) {
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group})
	if err != nil {
		return "", err
	}

	if listenHandler != nil {
		if err := ConfigClient.ListenConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
			OnChange: func(namespace, group, dataId, data string) {
				listenHandler(&data, nil)
			},
		}); err != nil {
			listenHandler(nil, err)
		}
	}
	return content, nil
}

// 获取配置项列表
func GetConfigList(page, count int) (*model.ConfigPage, error) {
	configPage, err := ConfigClient.SearchConfig(vo.SearchConfigParam{
		Search:   "blur",
		DataId:   "",
		Group:    "",
		PageNo:   page,
		PageSize: count,
	})

	if err != nil {
		return nil, err
	}
	// for _, page_item := range configPage.PageItems {
	// 	fmt.Println(page_item.DataId, page_item.Group, page_item.Tenant)
	// }
	return configPage, nil
}

// 设置配置
func SetConfig(dataId, group, content string) error {
	if ConfigClient == nil {
		return errors.New("nacos service is not connected. Please check the config file.")
	}

	success, err := ConfigClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})

	if err != nil {
		return err
	}

	if !success {
		return errors.New("Set Failed.")
	}

	return nil
}

// 订阅配置，当监听配置出现变化时给于通知
func ListenConfig(dataId, group string) <-chan string {
	updated_conf := make(chan string)
	err := ConfigClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			updated_conf <- data
		},
	})

	if err != nil {
		return nil
	}

	return updated_conf
}
