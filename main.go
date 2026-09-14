package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
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

// zoneID resolves the Designate zone that holds the challenge's FQDN.
func (c *designateDNSProviderSolver) zoneID(caller, resolvedZone string) (string, error) {
	allPages, err := zones.List(c.client, zones.ListOpts{Name: resolvedZone}).AllPages()
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract zones: %w", err)
	}

	if len(allZones) != 1 {
		return "", fmt.Errorf("%s: Expected to find 1 zone %s, found %v", caller, resolvedZone, len(allZones))
	}

	return allZones[0].ID, nil
}

// findTXTRecordSet returns the TXT recordset for fqdn, or nil when none exists.
func (c *designateDNSProviderSolver) findTXTRecordSet(zoneID, fqdn string) (*recordsets.RecordSet, error) {
	allPages, err := recordsets.ListByZone(c.client, zoneID, recordsets.ListOpts{
		Name: fqdn,
		Type: "TXT",
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list record sets: %w", err)
	}

	allRRs, err := recordsets.ExtractRecordSets(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract record sets: %w", err)
	}

	if len(allRRs) == 0 {
		return nil, nil
	}

	if len(allRRs) != 1 {
		return nil, fmt.Errorf("expected at most 1 TXT recordset for %s, found %v", fqdn, len(allRRs))
	}

	return &allRRs[0], nil
}

func (c *designateDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	log.Debugf("Present() called ch.DNSName=%s ch.ResolvedZone=%s ch.ResolvedFQDN=%s ch.Type=%s", ch.DNSName, ch.ResolvedZone, ch.ResolvedFQDN, ch.Type)

	zoneID, err := c.zoneID("Present", ch.ResolvedZone)
	if err != nil {
		return err
	}

	rs, err := c.findTXTRecordSet(zoneID, ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	record := quoteRecord(ch.Key)

	if rs == nil {
		_, err = recordsets.Create(c.client, zoneID, recordsets.CreateOpts{
			Name:    ch.ResolvedFQDN,
			Type:    "TXT",
			Records: []string{record},
			TTL:     60,
		}).Extract()
		if err != nil {
			return fmt.Errorf("failed to create record set: %w", err)
		}
		return nil
	}

	// A recordset already exists at this name. That happens whenever two
	// authorizations share one _acme-challenge name, most commonly a certificate
	// covering both example.com and *.example.com: both present a different value
	// at _acme-challenge.example.com. Creating a second recordset is not possible,
	// so add this challenge's record to the existing one.
	if slices.Contains(rs.Records, record) {
		log.Debugf("Present(): record already present in recordset %s", rs.ID)
		return nil
	}

	_, err = recordsets.Update(c.client, zoneID, rs.ID, recordsets.UpdateOpts{
		Records: append(slices.Clone(rs.Records), record),
	}).Extract()
	if err != nil {
		return fmt.Errorf("failed to add record to record set: %w", err)
	}

	return nil
}

func (c *designateDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	log.Debugf("CleanUp called ch.ResolvedZone=%s ch.ResolvedFQDN=%s", ch.ResolvedZone, ch.ResolvedFQDN)

	zoneID, err := c.zoneID("CleanUp", ch.ResolvedZone)
	if err != nil {
		return err
	}

	rs, err := c.findTXTRecordSet(zoneID, ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	// cleanup may happen multiple times, at least in tests
	// => no error if RS was already deleted
	if rs == nil {
		return nil
	}

	record := quoteRecord(ch.Key)
	remaining := make([]string, 0, len(rs.Records))
	for _, r := range rs.Records {
		if r != record {
			remaining = append(remaining, r)
		}
	}

	if len(remaining) == len(rs.Records) {
		log.Debugf("CleanUp(): record not present in recordset %s, nothing to do", rs.ID)
		return nil
	}

	// Remove the whole recordset only once this challenge's record was the last one
	// in it. Deleting it unconditionally would take a concurrent challenge's record
	// with it and leave that challenge waiting for a record that no longer exists.
	if len(remaining) == 0 {
		if err := recordsets.Delete(c.client, zoneID, rs.ID).ExtractErr(); err != nil {
			return fmt.Errorf("failed to delete recordset: %w", err)
		}
		return nil
	}

	_, err = recordsets.Update(c.client, zoneID, rs.ID, recordsets.UpdateOpts{
		Records: remaining,
	}).Extract()
	if err != nil {
		return fmt.Errorf("failed to remove record from record set: %w", err)
	}

	return nil
}

func (c *designateDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Debugf("Initialize called")

	cl, err := createDesignateServiceClient()
	if err != nil {
		return fmt.Errorf("failed to create designate client: %w", err)
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
