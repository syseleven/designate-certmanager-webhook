package main

import (
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

func TestRunsSuite(t *testing.T) {
	zone := os.Getenv("TEST_ZONE_NAME")
	if zone == "" {
		t.Fatalf("TEST_ZONE_NAME is not set")
	}

	solver := &designateDNSProviderSolver{}
	fixture := acmetest.NewFixture(solver,
		acmetest.SetResolvedZone(zone),
		acmetest.SetDNSName("*."+zone),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetConfig("{}"),
	)

	fixture.RunBasic(t)
	fixture.RunExtended(t)
}
