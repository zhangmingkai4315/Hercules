package utils

import (
	"testing"

	"github.com/prometheus/prometheus/util/testutil"
)

func TestGetFederationHostsFromConfig(t *testing.T) {
	feds, err := GetFederationHostsFromConfig("testdata/prometheus.conf")
	testutil.Ok(t, err)
	expectFeds := []string{"source-prometheus-1:9090", "source-prometheus-2:9090", "source-prometheus-3:9090"}
	testutil.Equals(t, expectFeds, feds)

	feds, err = GetFederationHostsFromConfig("testdata/prometheus.empty.conf")
	testutil.Ok(t, err)
	expectFeds = []string{}
	testutil.Equals(t, expectFeds, feds)

}
