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

var configClient config_client.IConfigClient

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

func init() {
	f_config, err := os.Open("config/config")
	if err != nil {
		panic(err)
	}

	defer f_config.Close()

	fileinfo, err := f_config.Stat()
	if err != nil {
		panic(err)
	}
	filesize := fileinfo.Size()
	data_buf := make([]byte, filesize)
	_, err = f_config.Read(data_buf)
	if err != nil {
		panic(err)
	}

	conf_data := ConfRoot{}
	err = json.Unmarshal(data_buf, &conf_data)
	if err != nil {
		panic(err)
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

	configClient, _ = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
}

func GetString(dataId string, group string) (string, error) {
	if configClient == nil {
		return "", errors.New("nacos service is not connected. Please check the config file.")
	}

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group})
	if err != nil {
		return "", err
	}

	return content, nil
}

func GetConfigList(page, count int) (*model.ConfigPage, error) {
	if configClient == nil {
		return nil, errors.New("nacos service is not connected. Please check the config file.")
	}

	configPage, err := configClient.SearchConfig(vo.SearchConfigParam{
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

func SetConfig(dataId, group, content string) error {
	if configClient == nil {
		return errors.New("nacos service is not connected. Please check the config file.")
	}

	success, err := configClient.PublishConfig(vo.ConfigParam{
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
