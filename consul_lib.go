package ginlib

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"strings"

	capi "github.com/hashicorp/consul/api"
)

// ConsulRegister 注册到consul; prometheus 是否暴露普罗米修斯的metrics接口
func ConsulRegister(r *gin.Engine, prometheus bool) {
	consulAddr := Ini_Str("app.consul_addr")
	if consulAddr == "" {
		return
	}
	localIP, err := LocalIP()
	if err != nil {
		Logger.Error("注册consul服务", zap.Error(err))
		return
	}
	//本地ip拦截
	ipFilter := func(ctx *gin.Context) {
		if strings.Index(ctx.Request.Host, localIP) == -1 {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
	}
	//添加健康检查路由
	r.GET("/health", ipFilter, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "success")
		return
	})
	//添加普罗米修斯路由
	if prometheus {
		r.GET("/metrics", ipFilter, gin.WrapH(promhttp.Handler()))
	}
	//注册consul服务
	consulConfig := capi.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := capi.NewClient(consulConfig)
	if err != nil {
		Logger.Error("注册consul服务", zap.Error(err))
		return
	}
	port, _ := strconv.Atoi(APP_PORT)
	tags := []string{"app=" + APP_NAME, "env=" + GetEnv()}
	if prometheus {
		tags = append(tags, "prometheus=true")
	}
	registration := &capi.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s:%d", localIP, port),
		Name:    "app-" + APP_NAME,
		Port:    port,
		Address: localIP,
		Tags:    tags,
		Check: &capi.AgentServiceCheck{
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "60s",
			HTTP:                           fmt.Sprintf("http://%s:%s/health", localIP, APP_PORT),
		},
	}
	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		Logger.Error("注册consul服务", zap.Error(err))
		return
	}
}

// LocalIP 获取本地ip
func LocalIP() (ip string, err error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}
	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 排除回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			// 获取接口的IP地址
			var addrs []net.Addr
			addrs, err = iface.Addrs()
			if err != nil {
				continue
			}

			// 遍历接口的IP地址
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					return ipNet.IP.String(), nil
				}
			}
		}
	}
	return
}
