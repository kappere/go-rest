package rest

import (
	"embed"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AppName    string `yaml:"AppName"`
	Port       int    `yaml:"Port"`
	ConfigFile string `yaml:"ConfigFile"`
	Debug      bool   `yaml:"Debug"`
	Profile    string `yaml:"Profile"`

	Log            LogConfig            `yaml:"Log"`
	Database       DatabaseConfig       `yaml:"Database"`
	Redis          RedisConfig          `yaml:"Redis"`
	Session        SessionConfig        `yaml:"Session"`
	PeriodLimit    PeriodLimitConfig    `yaml:"PeriodLimit"`
	OAuth2         OAuth2Config         `yaml:"OAuth2"`
	StaticResource StaticResourceConfig `yaml:"StaticResource"`
	Rpc            RpcConfig            `yaml:"Rpc"`
}

type LogConfig struct {
	Path     string `yaml:"Path"`
	Level    string `yaml:"Level"`
	SaveDays int    `yaml:"SaveDays"`
}

type DatabaseConfig struct {
	Dialector interface{}
	Type      string `yaml:"Type"`
	Dsn       string `yaml:"Dsn"`
	LogLevel  int    `yaml:"LogLevel"`
}

type RedisConfig struct {
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
}

type SessionConfig struct {
	Name      string        `yaml:"Name"`
	Domain    string        `yaml:"Domain"`
	Path      string        `yaml:"Path"`
	Secure    bool          `yaml:"Secure"`
	HttpOnly  bool          `yaml:"HttpOnly"`
	MaxAge    int           `yaml:"MaxAge"`
	SameSite  http.SameSite `yaml:"SameSite"`
	StoreType string        `yaml:"StoreType"`
}

type PeriodLimitConfig struct {
	Enable bool `yaml:"Enable"`
	Period int  `yaml:"Period"`
	Quota  int  `yaml:"Quota"`
}

type OAuth2Config struct {
	Enable   bool   `yaml:"Enable"`
	Expire   int    `yaml:"Expire"`
	TokenUri string `yaml:"TokenUri"`
}

type RpcConfig struct {
	Token      string           `yaml:"Token"`
	Type       string           `yaml:"Type"`
	IpProxy    IpProxyConfig    `yaml:"IpProxy"`
	Kubernetes KubernetesConfig `yaml:"Kubernetes"`
}

type IpProxyConfig struct {
	Proxy map[string]string `yaml:"Proxy"`
}

type KubernetesConfig struct {
	Namespace string            `yaml:"Namespace"`
	Proxy     map[string]string `yaml:"Proxy"`
	PortName  string            `yaml:"PortName"`
}

type StaticResourceConfig struct {
	Location string
	Fs       interface{}
}

type MockConfig struct {
	Config `yaml:"Config"`
}

var defaultConf = MockConfig{
	Config{
		AppName:    "app",
		Port:       80,
		ConfigFile: "",
		Debug:      false,
		Profile:    "prod",
		Log: LogConfig{
			Path:     "log/app",
			SaveDays: -1,
		},
		Session: SessionConfig{
			Name:      "sessionid",
			Domain:    "",
			Path:      "/",
			Secure:    false,
			HttpOnly:  true,
			MaxAge:    86400 * 30,
			SameSite:  http.SameSiteDefaultMode,
			StoreType: CACHE_TYPE_MEMORY,
		},
		PeriodLimit: PeriodLimitConfig{
			Enable: false,
			Period: 5,
			Quota:  100,
		},
		OAuth2: OAuth2Config{
			Enable:   false,
			Expire:   7200,
			TokenUri: "/token",
		},
		Database: DatabaseConfig{
			// 0-panic 1-fatal 2-error 3-warn 4-info 5-debug
			LogLevel: 4,
		},
		Rpc: RpcConfig{
			Token: strconv.Itoa(rand.Int()),
			// Kubernetes IpProxy
			Type: "IpProxy",
			IpProxy: IpProxyConfig{
				Proxy: map[string]string{
					"*": "http://127.0.0.1:8080",
				},
			},
			Kubernetes: KubernetesConfig{
				Namespace: "default",
				Proxy: map[string]string{
					"*": "http://127.0.0.1:8080/api/v1/namespaces/{namespace}/services/http:{app}:/proxy",
				},
				PortName: "http",
			},
		},
	},
}

func LoadConfig(args []string, targetConf interface{}, configFs embed.FS) {
	port := flag.Int("port", 0, "port")
	pConfigFile := flag.String("config", "", "config file")
	pProfile := flag.String("profile", "", "profile")
	flag.Parse()

	var currentConfig MockConfig

	// 读取默认配置
	defaultConfigJsonByte, err := yaml.Marshal(defaultConf)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(defaultConfigJsonByte, targetConf)

	// 读取内部文件配置
	internalConfigBytes, err := configFs.ReadFile("config.yaml")
	if err == nil {
		yaml.Unmarshal(internalConfigBytes, targetConf)
	}

	// 读取profile
	copyConfig(targetConf, &currentConfig)
	if *pProfile != "" {
		currentConfig.Profile = *pProfile
		mergeConfig(&currentConfig, targetConf)
	}
	profile := currentConfig.Profile

	if profile != "" {
		// 读取内部profile文件配置
		internalConfigBytes, err = configFs.ReadFile("config-" + profile + ".yaml")
		if err != nil {
			panic(err)
		}
		yaml.Unmarshal(internalConfigBytes, targetConf)
	}

	// 读取外部配置
	copyConfig(targetConf, &currentConfig)
	if currentConfig.ConfigFile == "" {
		// 未配置外部文件，默认取config.yaml
		if _, err := os.Stat("config.yaml"); err == nil {
			currentConfig.ConfigFile = "config.yaml"
		}
	}
	if *pConfigFile != "" {
		currentConfig.ConfigFile = *pConfigFile
		mergeConfig(&currentConfig, targetConf)
	}
	configFile := currentConfig.ConfigFile
	if configFile != "" {
		if !strings.HasSuffix(configFile, ".yaml") {
			panic("yaml config file needed")
		}
		fileByte, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic(err)
		}
		yaml.Unmarshal(fileByte, targetConf)
	}

	// 读取命令行配置
	if *port > 0 {
		copyConfig(targetConf, &currentConfig)
		currentConfig.Port = *port
		mergeConfig(&currentConfig, targetConf)
	}
}

func copyConfig(conf interface{}, mockConf *MockConfig) {
	mockConfigYaml, _ := yaml.Marshal(conf)
	yaml.Unmarshal(mockConfigYaml, mockConf)
}

func mergeConfig(mockConf *MockConfig, conf interface{}) {
	mockConfigYaml, _ := yaml.Marshal(mockConf)
	yaml.Unmarshal(mockConfigYaml, conf)
}
