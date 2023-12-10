package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/httpx"
)

var rpcConf conf.RpcConfig

var srvLookup func(srvname string) RpcService

type RpcService struct {
	Name string
	Addr string
}

type RpcResult struct {
	Data []byte
	Err  error
}

func (r RpcResult) ToMap() (interface{}, error) {
	resp := httpx.Ok(nil)
	err := json.Unmarshal(r.Data, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return resp.GetData(), nil
}

func (r RpcResult) ToObj(result interface{}) error {
	data, err := r.ToMap()
	if err != nil {
		return err
	}
	objByteData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(objByteData, result)
}

func InitClient(c conf.RpcConfig) {
	rpcConf = c
	slog.Info("Init rpc client,", "type", rpcConf.Type)
	if strings.ToLower(rpcConf.Type) == "kubernetes" {
		if isInKubernetesCluster() {
			slog.Info("In kubernetes")
			// minikube需要先添加service读取权限
			// kubectl create clusterrolebinding service-reader-pod --clusterrole=service-reader --serviceaccount=default:default
			srvLookup = func(srvname string) RpcService {
				_, addrs, _ := net.LookupSRV(rpcConf.Kubernetes.PortName, "tcp", srvname)
				if len(addrs) > 0 {
					addr := "http://" + addrs[0].Target + ":" + strconv.FormatInt(int64(addrs[0].Port), 10)
					return RpcService{
						Name: srvname,
						Addr: addr,
					}
				}
				return RpcService{Name: srvname}
			}
		} else {
			slog.Info("Out of kubernetes")
			defaultProxyAddr := rpcConf.Kubernetes.Proxy["*"]
			srvLookup = func(srvname string) RpcService {
				addr := rpcConf.Kubernetes.Proxy[srvname]
				if addr == "" {
					addr = defaultProxyAddr
				}
				addr = strings.ReplaceAll(addr, "{namespace}", rpcConf.Kubernetes.Namespace)
				addr = strings.ReplaceAll(addr, "{app}", srvname)
				return RpcService{
					Name: srvname,
					Addr: addr,
				}
			}
		}
	} else if strings.ToLower(rpcConf.Type) == "ipproxy" {
		defaultProxyAddr := rpcConf.IpProxy.Proxy["*"]
		srvLookup = func(srvname string) RpcService {
			addr := rpcConf.IpProxy.Proxy[srvname]
			if addr == "" {
				addr = defaultProxyAddr
			}
			return RpcService{
				Name: srvname,
				Addr: addr,
			}
		}
	}
}

func (service RpcService) Call(url string, body map[string]interface{}) RpcResult {
	data, err := httpPost(service.Addr+RPC_PREFIX+url, body)
	if err != nil {
		return RpcResult{nil, err}
	}
	return RpcResult{data, nil}
}

func httpPost(url string, body map[string]interface{}) ([]byte, error) {
	reqbody := strings.NewReader("")
	if body != nil {
		jsonbody, _ := json.Marshal(body)
		reqbody = strings.NewReader(string(jsonbody))
	}
	request, err := http.NewRequest("POST", url, reqbody)
	if err != nil {
		return nil, err
	}
	return apply(request)
}

func apply(request *http.Request) ([]byte, error) {
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
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func Service(srvname string) RpcService {
	srv := srvLookup(srvname)
	if srv.Addr == "" {
		panic("Service not found: " + srvname)
	}
	return srv
}

func isInKubernetesCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}
