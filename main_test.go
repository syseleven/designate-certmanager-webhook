package main

import (
	"os"
	"testing"

	logf "github.com/cert-manager/cert-manager/pkg/logs"
	acmetest "github.com/cert-manager/cert-manager/test/acme"

	ctrl "sigs.k8s.io/controller-runtime"
)

func TestRunsSuite(t *testing.T) {
	logf.InitLogs()
	defer logf.FlushLogs()
	ctrl.SetLogger(logf.Log)

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
