package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/rest"
)

var rpcConf *rest.RpcConfig

var srvLookup func(srvname string) *RpcService

type RpcService struct {
	Name string
	Addr string
}

func InitClient(c *rest.RpcConfig) {
	rpcConf = c

	logger.Info("service type: %s", rpcConf.Type)
	if rpcConf.Type == "Kubernetes" {
		if isInKubernetesCluster() {
			logger.Info("in kubernetes")
			// minikube需要先添加service读取权限
			// kubectl create clusterrolebinding service-reader-pod --clusterrole=service-reader --serviceaccount=default:default
			srvLookup = func(srvname string) *RpcService {
				_, addrs, _ := net.LookupSRV(rpcConf.Kubernetes.PortName, "tcp", srvname)
				if len(addrs) > 0 {
					addr := "http://" + addrs[0].Target + ":" + strconv.FormatInt(int64(addrs[0].Port), 10)
					return &RpcService{
						Name: srvname,
						Addr: addr,
					}
				}
				return nil
			}
		} else {
			logger.Info("out of kubernetes")
			defaultProxyAddr := rpcConf.Kubernetes.Proxy["*"]
			srvLookup = func(srvname string) *RpcService {
				addr := rpcConf.Kubernetes.Proxy[srvname]
				if addr == "" {
					addr = defaultProxyAddr
				}
				addr = strings.ReplaceAll(addr, "{namespace}", rpcConf.Kubernetes.Namespace)
				addr = strings.ReplaceAll(addr, "{app}", srvname)
				return &RpcService{
					Name: srvname,
					Addr: addr,
				}
			}
		}
	} else if rpcConf.Type == "IpProxy" {
		defaultProxyAddr := rpcConf.IpProxy.Proxy["*"]
		srvLookup = func(srvname string) *RpcService {
			addr := rpcConf.IpProxy.Proxy[srvname]
			if addr == "" {
				addr = defaultProxyAddr
			}
			return &RpcService{
				Name: srvname,
				Addr: addr,
			}
		}
	}
}

func (service *RpcService) Call(url string, body map[string]interface{}) interface{} {
	return httpPost(service.Addr+RPC_PREFIX+url, body)
}

// func httpGet(url string) interface{} {
// 	request, _ := http.NewRequest("GET", url, nil)
// 	return apply(request)
// }

func httpPost(url string, body map[string]interface{}) interface{} {
	reqbody := strings.NewReader("")
	if body != nil {
		jsonbody, _ := json.Marshal(body)
		reqbody = strings.NewReader(string(jsonbody))
	}
	reqest, _ := http.NewRequest("POST", url, reqbody)
	return apply(reqest)
}

func apply(request *http.Request) interface{} {
	client := &http.Client{}

	// 计算token
	if rpcConf.Token != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		hash := sha256.New()
		randstr := strconv.Itoa(rand.Int())
		hash.Write([]byte(rpcConf.Token + "#" + randstr + "#" + timestamp))
		enc := hex.EncodeToString(hash.Sum(nil))
		request.Header.Add("inner_token_enc", enc+"#"+randstr+"#"+timestamp)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	bytedata, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	resp := rest.Resp{}
	json.Unmarshal(bytedata, &resp)
	if !resp.Success {
		panic(resp.Message)
	}
	return resp.Data
}

func Service(srvname string) *RpcService {
	srv := srvLookup(srvname)
	if srv == nil {
		panic("service not found: " + srvname)
	}
	return srv
}

func isInKubernetesCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}
