package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zhangmingkai4315/hercules/handlers"
	"github.com/zhangmingkai4315/hercules/utils"
)

var (
	prometheusConfig      string
	currentPrometheusHost string
	agentPort             int
	logLevel              string
)

func init() {
	flag.StringVar(&prometheusConfig, "prometheus.config", "", "current prometheus config file")
	flag.StringVar(&currentPrometheusHost, "prometheus.host", "", "current prometheus host and port (for federation)")
	flag.IntVar(&agentPort, "agent.port", 19090, "Agent port for connect with other")
	flag.StringVar(&logLevel, "log.level", "warning", "Setting log level for program")
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
func setLogLevel(level string) {
	log.SetOutput(os.Stdout)
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if logLevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if logLevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	flag.Parse()
	setLogLevel(logLevel)
	if prometheusConfig == "" || currentPrometheusHost == "" {
		flag.Usage()
		os.Exit(1)
	}
	node, err := utils.NewPrometheusNode(currentPrometheusHost)
	checkError(err)
	feds, err := utils.GetFederationHostsFromConfig(prometheusConfig)
	checkError(err)
	node.Children = utils.NewPrometheusNodeList(feds)
	quit := make(chan struct{})
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				node.Ping()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.HandleFunc("/status", handlers.HealthCheckHandler)
	http.HandleFunc("/proxy", handlers.RequestProxy)
	http.HandleFunc("/graph", utils.GetGraph(node))
	http.HandleFunc("/update-graph", utils.UpdateGraph(node))
	http.ListenAndServe(fmt.Sprintf(":%d", agentPort), nil)
}
