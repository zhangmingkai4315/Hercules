package utils

import (
	"fmt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
)

// GetFederationHostsFromConfig read prometheus config file
// and return all target for federation, only static_config support
func GetFederationHostsFromConfig(path string) ([]string, error) {
	federations := []string{}
	conf, err := config.LoadFile(path)
	if err != nil {
		return federations, err
	}
	for _, sc := range conf.ScrapeConfigs {
		if sc.MetricsPath == "/federate" {
			// federations = append(federations,sc.)
			for _, tg := range sc.ServiceDiscoveryConfig.StaticConfigs {
				for _, t := range tg.Targets {
					fmt.Println(t)
					if err == config.CheckTargetAddress(t[model.AddressLabel]) {
						federations = append(federations, string(t[model.AddressLabel]))
					}
				}
			}
		}
	}
	return federations, nil
}
