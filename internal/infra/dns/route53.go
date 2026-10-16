package dns

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/validation"
)

const route53Base = "https://route53.amazonaws.com"

type Route53Provider struct {
	accessKey string
	secretKey string
	session   string
	client    *http.Client
	baseURL   string
	now       func() time.Time
}

func NewRoute53Provider(accessKey, secretKey, sessionToken string, client *http.Client) (*Route53Provider, error) {
	if strings.TrimSpace(accessKey) == "" || strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("route53 access key and secret are required")
	}
	if client == nil {
		client = validation.NewSafeHTTPClient(30 * time.Second)
	}
	return &Route53Provider{accessKey: accessKey, secretKey: secretKey, session: sessionToken, client: client, baseURL: route53Base, now: time.Now}, nil
}

func (p *Route53Provider) Name() string { return "route53" }

type route53ZoneList struct {
	XMLName xml.Name `xml:"ListHostedZonesResponse"`
	Zones   []struct {
		ID     string `xml:"Id"`
		Name   string `xml:"Name"`
		Config struct {
			PrivateZone bool `xml:"PrivateZone"`
		} `xml:"Config"`
	} `xml:"HostedZones>HostedZone"`
	Truncated bool `xml:"IsTruncated"`
}

func (p *Route53Provider) zoneID(ctx context.Context, fqdn string) (string, error) {
	name := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(name) {
		return "", fmt.Errorf("dns record name is required")
	}
	raw, err := p.call(ctx, http.MethodGet, "/2013-04-01/hostedzone", "", nil)
	if err != nil {
		return "", err
	}
	var zones route53ZoneList
	if err := xml.Unmarshal(raw, &zones); err != nil {
		return "", fmt.Errorf("decode route53 zones: %w", err)
	}
	if zones.Truncated {
		return "", fmt.Errorf("route53 zone list is truncated; narrow the account zones")
	}
	best := ""
	bestLen := -1
	for _, zone := range zones.Zones {
		if zone.Config.PrivateZone {
			continue
		}
		zoneName := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(zone.Name)), ".")
		if name != zoneName && !strings.HasSuffix(name, "."+zoneName) {
			continue
		}
		if len(zoneName) <= bestLen {
			continue
		}
		best = strings.TrimPrefix(zone.ID, "/hostedzone/")
		bestLen = len(zoneName)
	}
	if best == "" {
		return "", fmt.Errorf("no route53 zone covers %s", fqdn)
	}
	return best, nil
}

type route53RecordSet struct {
	XMLName xml.Name `xml:"ResourceRecordSet"`
	Name    string   `xml:"Name"`
	Type    string   `xml:"Type"`
	TTL     int      `xml:"TTL"`
	Values  []string `xml:"ResourceRecords>ResourceRecord>Value"`
}

type route53RecordSets struct {
	XMLName xml.Name           `xml:"ListResourceRecordSetsResponse"`
	Sets    []route53RecordSet `xml:"ResourceRecordSets>ResourceRecordSet"`
}

func quoteTXT(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func unquoteTXT(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		value = value[1 : len(value)-1]
	}
	return strings.ReplaceAll(value, `\"`, `"`)
}

func (p *Route53Provider) readTXTSet(ctx context.Context, zoneID, fqdn string) ([]string, error) {
	raw, err := p.call(ctx, http.MethodGet, "/2013-04-01/hostedzone/"+zoneID+"/rrset", "type=TXT&name="+fqdn, nil)
	if err != nil {
		return nil, err
	}
	var sets route53RecordSets
	if err := xml.Unmarshal(raw, &sets); err != nil {
		return nil, fmt.Errorf("decode route53 records: %w", err)
	}
	for _, set := range sets.Sets {
		if set.Type != "TXT" || !strings.EqualFold(strings.TrimSuffix(set.Name, "."), fqdn) {
			continue
		}
		values := make([]string, 0, len(set.Values))
		for _, value := range set.Values {
			values = append(values, unquoteTXT(value))
		}
		sort.Strings(values)
		return values, nil
	}
	return nil, nil
}

func (p *Route53Provider) writeTXTSet(ctx context.Context, zoneID, fqdn string, values []string, action string) error {
	type record struct {
		XMLName xml.Name `xml:"ResourceRecord"`
		Value   string   `xml:"Value"`
	}
	type recordSet struct {
		XMLName xml.Name `xml:"ResourceRecordSet"`
		Name    string   `xml:"Name"`
		Type    string   `xml:"Type"`
		TTL     int      `xml:"TTL"`
		Records []record `xml:"ResourceRecords>ResourceRecord"`
	}
	type change struct {
		XMLName xml.Name  `xml:"Change"`
		Action  string    `xml:"Action"`
		Set     recordSet `xml:"ResourceRecordSet"`
	}
	type batch struct {
		XMLName xml.Name `xml:"ChangeResourceRecordSetsRequest"`
		Xmlns   string   `xml:"xmlns,attr"`
		Changes []change `xml:"ChangeBatch>Changes>Change"`
	}
	set := recordSet{Name: fqdn + ".", Type: "TXT", TTL: 120}
	for _, value := range values {
		set.Records = append(set.Records, record{Value: quoteTXT(value)})
	}
	body, err := xml.Marshal(batch{Xmlns: "https://route53.amazonaws.com/doc/2013-04-01/", Changes: []change{{Action: action, Set: set}}})
	if err != nil {
		return fmt.Errorf("encode route53 change: %w", err)
	}
	_, err = p.call(ctx, http.MethodPost, "/2013-04-01/hostedzone/"+zoneID+"/rrset", "", append([]byte(xml.Header), body...))
	return err
}

func (p *Route53Provider) Present(ctx context.Context, fqdn, value string) error {
	fqdn = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(fqdn) || value == "" || len(value) > 1024 {
		return fmt.Errorf("dns record name and value are required")
	}
	zone, err := p.zoneID(ctx, fqdn)
	if err != nil {
		return err
	}
	existing, err := p.readTXTSet(ctx, zone, fqdn)
	if err != nil {
		return err
	}
	for _, current := range existing {
		if current == value {
			return nil
		}
	}
	return p.writeTXTSet(ctx, zone, fqdn, append(existing, value), "UPSERT")
}

func (p *Route53Provider) CleanupValue(ctx context.Context, fqdn, value string) error {
	fqdn = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(fqdn) {
		return fmt.Errorf("dns record name is required")
	}
	zone, err := p.zoneID(ctx, fqdn)
	if err != nil {
		return err
	}
	existing, err := p.readTXTSet(ctx, zone, fqdn)
	if err != nil {
		return err
	}
	remaining := make([]string, 0, len(existing))
	for _, current := range existing {
		if current != value {
			remaining = append(remaining, current)
		}
	}
	if len(remaining) == len(existing) {
		return nil
	}
	if len(remaining) == 0 {
		return p.writeTXTSet(ctx, zone, fqdn, existing, "DELETE")
	}
	return p.writeTXTSet(ctx, zone, fqdn, remaining, "UPSERT")
}

func (p *Route53Provider) Cleanup(ctx context.Context, fqdn string) error {
	fqdn = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(fqdn) {
		return fmt.Errorf("dns record name is required")
	}
	zone, err := p.zoneID(ctx, fqdn)
	if err != nil {
		return err
	}
	existing, err := p.readTXTSet(ctx, zone, fqdn)
	if err != nil {
		return err
	}
	if len(existing) == 0 {
		return nil
	}
	return p.writeTXTSet(ctx, zone, fqdn, existing, "DELETE")
}
