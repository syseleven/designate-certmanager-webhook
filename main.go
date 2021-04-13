package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
)

const GroupName = "acme.syseleven.de"

func main() {
	cmd.RunWebhookServer(GroupName,
		&designateDNSProviderSolver{},
	)
}

type designateDNSProviderSolver struct {
	client *gophercloud.ServiceClient
}

func (c *designateDNSProviderSolver) Name() string {
	return "designateDNS"
}

func (c *designateDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	log.Debugf("Present() called ch.DNSName=%s ch.ResolvedZone=%s ch.ResolvedFQDN=%s ch.Type=%s", ch.DNSName, ch.ResolvedZone, ch.ResolvedFQDN, ch.Type)

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
	opts.Type = "TXT"
	opts.Records = []string{quoteRecord(ch.Key)}

	_, err = recordsets.Create(c.client, allZones[0].ID, opts).Extract()
	if err != nil {
		return err
	}

	return nil
}

func (c *designateDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	log.Debugf("CleanUp called ch.ResolvedZone=%s ch.ResolvedFQDN=%s", ch.ResolvedZone, ch.ResolvedFQDN)

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
		Data: quoteRecord(ch.Key),
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

func (c *designateDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Debugf("Initialize called")

	cl, err := createDesignateServiceClient()
	if err != nil {
		return err
	}

	c.client = cl
	return nil
}

func quoteRecord(r string) string {
	if strings.HasPrefix(r, "\"") && strings.HasSuffix(r, "\"") {
		return r
	} else {
		return strconv.Quote(r)
	}
}
