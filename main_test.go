package main

import (
	"os"
	"testing"

	//"github.com/cert-manager/cert-manager/test/acme/dns"
	dns "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	solver := New()
	fixture := dns.NewFixture(solver,
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/my-custom-solver"),
	)

	fixture.RunBasic(t)
	fixture.RunExtended(t)
}
