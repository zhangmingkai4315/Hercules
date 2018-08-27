package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/hercules/handlers"
	"github.com/zhangmingkai4315/hercules/utils"
)

var (
	prometheusConfig      string
	currentPrometheusHost string
	agentPort             int
)

func init() {
	flag.StringVar(&prometheusConfig, "prometheus.config", "", "current prometheus config file")
	flag.StringVar(&currentPrometheusHost, "prometheus.host", "", "current prometheus host and port (for federation)")
	flag.IntVar(&agentPort, "agent.port", 19090, "Agent port for connect with other")
}
func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
func main() {
	flag.Parse()
	if prometheusConfig == "" || currentPrometheusHost == "" {
		flag.Usage()
		os.Exit(1)
	}
	node, err := utils.NewPrometheusNode(currentPrometheusHost)
	checkError(err)
	feds, err := utils.GetFederationHostsFromConfig(prometheusConfig)
	checkError(err)
	node.Children = utils.NewPrometheusNodeList(feds)

	http.HandleFunc("/status", handlers.HealthCheckHandler)
	http.HandleFunc("/proxy", handlers.RequestProxy)

	http.ListenAndServe(fmt.Sprintf(":%d", agentPort), nil)

}
