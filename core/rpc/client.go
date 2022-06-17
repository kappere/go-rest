package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
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

var srvMap = make(map[string]*RpcService)

type RpcService struct {
	Name string
	Addr string
}

func InitClient(c *rest.RpcConfig) {
	rpcConf = c

	if rpcConf.Type == "kubernetes" {
		if isInKubernetesCluster() {
			logger.Info("in kubernetes")
			// 需要先添加service读取权限
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
			srvLookup = func(srvname string) *RpcService {
				addr := rpcConf.Kubernetes.Proxy[srvname]
				if addr == "" {
					addr = rpcConf.Kubernetes.Proxy["*"]
				}
				addr = strings.ReplaceAll(addr, "{namespace}", rpcConf.Kubernetes.Namespace)
				addr = strings.ReplaceAll(addr, "{app}", srvname)
				return &RpcService{
					Name: srvname,
					Addr: addr,
				}
			}
		}
	}
}

func (service *RpcService) Call(url string, body map[string]interface{}) interface{} {
	return httpPost(service.Addr+RPC_PREFIX+url, body)
}

func httpGet(url string) interface{} {
	reqest, _ := http.NewRequest("GET", url, nil)
	return apply(reqest)
}

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
	rpcToken := rpcConf.Token
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	hash := sha256.New()
	randstr := strconv.Itoa(rand.Int())
	hash.Write([]byte(rpcToken + "#" + randstr + "#" + timestamp))
	enc := hex.EncodeToString(hash.Sum(nil))

	request.Header.Add("inner_token_enc", enc+"#"+randstr+"#"+timestamp)
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	bytedata, _ := ioutil.ReadAll(response.Body)
	resp := rest.Resp{}
	json.Unmarshal(bytedata, &resp)
	if !resp.Success {
		panic(resp.Message)
	}
	return resp.Data
}

func Service(srvname string) *RpcService {
	return srvLookup(srvname)
}

func isInKubernetesCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}