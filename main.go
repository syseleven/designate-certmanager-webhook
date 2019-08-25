package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"github.com/kubernetes-incubator/external-dns/pkg/tlsutils"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&designateDNSProviderSolver{},
	)
}

// customDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type designateDNSProviderSolver struct {
	client *gophercloud.ServiceClient
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *designateDNSProviderSolver) Name() string {
	log.Infof("Name() called")
	return "designateDNS"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *designateDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	log.Infof("Present() called ch.DNSName=%s ch.ResolvedZone=%s ch.ResolvedFQDN=%s ch.Type=%s", ch.DNSName, ch.ResolvedZone, ch.ResolvedFQDN, ch.Type)

	listOpts := zones.ListOpts{
		Name: ch.ResolvedZone,
	}

	allPages, err := zones.List(c.client, listOpts).AllPages()
	if err != nil {
		return err
	}

	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return err
	}

	if len(allZones) != 1 {
		return fmt.Errorf("Present: Expected to find 1 zone %s, found %v", ch.ResolvedZone, len(allZones))
	}

	var opts recordsets.CreateOpts
	opts.Name = ch.ResolvedFQDN
	opts.Records = []string{ch.Key}
	opts.Type = "TXT"

	_, err = recordsets.Create(c.client, allZones[0].ID, opts).Extract()
	if err != nil {
		return err
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *designateDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	log.Infof("CleanUp called ch.ResolvedZone=%s ch.ResolvedFQDN=%s", ch.ResolvedZone, ch.ResolvedFQDN)

	listOpts := zones.ListOpts{
		Name: ch.ResolvedZone,
	}

	allPages, err := zones.List(c.client, listOpts).AllPages()
	if err != nil {
		return err
	}

	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return err
	}

	if len(allZones) != 1 {
		return fmt.Errorf("CleanUp: Expected to find 1 zone %s, found %v", ch.ResolvedZone, len(allZones))
	}

	recordListOpts := recordsets.ListOpts{
		Name: ch.ResolvedFQDN,
		Type: "TXT",
		Data: ch.Key,
	}

	allRecordPages, err := recordsets.ListByZone(c.client, allZones[0].ID, recordListOpts).AllPages()
	if err != nil {
		return err
	}

	allRRs, err := recordsets.ExtractRecordSets(allRecordPages)
	if err != nil {
		return err
	}

	if len(allRRs) != 1 {
		return fmt.Errorf("CleanUp: Expected to find 1 recordset matching %s in zone %s, found %v", ch.ResolvedFQDN, ch.ResolvedZone, len(allRRs))
	}

	// TODO rather than deleting the whole recordset we may have to delete individual records from it, i.e. perform an update rather than a delete
	err = recordsets.Delete(c.client, allZones[0].ID, allRRs[0].ID).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *designateDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Debugf("Initialize called")

	cl, err := createDesignateServiceClient()
	if err != nil {
		return err
	}

	c.client = cl
	return nil
}

// copies environment variables to new names without overwriting existing values
func remapEnv(mapping map[string]string) {
	for k, v := range mapping {
		currentVal := os.Getenv(k)
		newVal := os.Getenv(v)
		if currentVal == "" && newVal != "" {
			os.Setenv(k, newVal)
		}
	}
}

// returns OpenStack Keystone authentication settings by obtaining values from standard environment variables.
// also fixes incompatibilities between gophercloud implementation and *-stackrc files that can be downloaded
// from OpenStack dashboard in latest versions
func getAuthSettings() (gophercloud.AuthOptions, error) {
	remapEnv(map[string]string{
		"OS_TENANT_NAME": "OS_PROJECT_NAME",
		"OS_TENANT_ID":   "OS_PROJECT_ID",
		"OS_DOMAIN_NAME": "OS_USER_DOMAIN_NAME",
		"OS_DOMAIN_ID":   "OS_USER_DOMAIN_ID",
	})

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return gophercloud.AuthOptions{}, err
	}
	opts.AllowReauth = true
	if !strings.HasSuffix(opts.IdentityEndpoint, "/") {
		opts.IdentityEndpoint += "/"
	}
	if !strings.HasSuffix(opts.IdentityEndpoint, "/v2.0/") && !strings.HasSuffix(opts.IdentityEndpoint, "/v3/") {
		opts.IdentityEndpoint += "v2.0/"
	}
	return opts, nil
}

// authenticate in OpenStack and obtain Designate service endpoint
func createDesignateServiceClient() (*gophercloud.ServiceClient, error) {
	opts, err := getAuthSettings()
	if err != nil {
		return nil, err
	}
	log.Infof("Using OpenStack Keystone at %s", opts.IdentityEndpoint)
	authProvider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := tlsutils.CreateTLSConfig("OPENSTACK")
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}
	authProvider.HTTPClient.Transport = transport

	if err = openstack.Authenticate(authProvider, opts); err != nil {
		return nil, err
	}

	eo := gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	client, err := openstack.NewDNSV2(authProvider, eo)
	if err != nil {
		return nil, err
	}
	log.Infof("Found OpenStack Designate service at %s", client.Endpoint)
	return client, nil
}
