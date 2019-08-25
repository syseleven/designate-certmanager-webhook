package main

import (
	"os"
	"testing"
	"time"

	"github.com/jetstack/cert-manager/test/acme/dns"
)

var (
	zone    = os.Getenv("TEST_ZONE_NAME")
	binpath = os.Getenv("TEST_BINPATH")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	fixture := dns.NewFixture(&designateDNSProviderSolver{},
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetBinariesPath(binpath),
		dns.SetManifestPath("testdata/my-custom-solver"),
		dns.SetPropagationLimit(time.Minute*10),
	)

	fixture.RunConformance(t)
}
