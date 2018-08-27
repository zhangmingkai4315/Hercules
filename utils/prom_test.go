package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	currentNode = "source-prometheus:9090"
)

func TestGetFederationHostsFromConfig(t *testing.T) {
	feds, err := GetFederationHostsFromConfig("testdata/prometheus.conf")
	assert.Nil(t, err)
	expectFeds := []string{"source-prometheus-1:9090", "source-prometheus-2:9090", "source-prometheus-3:9090"}
	assert.ElementsMatch(t, expectFeds, feds)
	feds, err = GetFederationHostsFromConfig("testdata/prometheus.empty.conf")
	assert.Nil(t, err)
	expectFeds = []string{}
	assert.ElementsMatch(t, expectFeds, feds)
}
