package netutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrivateNetworkBlocklist(t *testing.T) {
	dialer := &net.Dialer{
		Control: PrivateNetworkBlocklist.Control,
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = dialer.DialContext
	client := &http.Client{
		Transport: transport,
	}

	tests := []struct {
		addr        string
		shouldBlock bool
	}{
		{"localhost", true},
		{"10.255.0.1", true},
		{"www.example.com", false},
		{"203.0.113.1", true},
		{"192.0.0.170", true},
		{"255.255.255.255", true},
		{"169.254.169.25", true},
		{"192.88.99.1", true},
		{"[::]", true},
		{"[::0]", true},
		{"[::1]", true},
		{"[2001:2::1]", true},
		{"[2001:4860:4802:32::a]", false},
		// TODO: Uncomment when Go 1.26 is released or backport is available
		// IPv4-mapped IPv6 addresses are rejected by net/url.Parse in Go 1.25.3 due to CVE-2025-47912 fix
		// This is fixed in Go 1.26: https://github.com/golang/go/issues/75815
		// {"[::ffff:192.168.0.1]", true},
	}

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+test.addr, nil)
		if err != nil {
			t.Fatalf("failed to create request for %s: %v", test.addr, err)
		}
		_, err = client.Do(req)
		var berr ErrBlocked
		if test.shouldBlock {
			assert.ErrorAs(t, err, &ErrBlocked{})
		} else {
			if errors.As(err, &berr) {
				t.Errorf("should not block %s but got error: %s", test.addr, err)
			}
		}
		cancel()
	}
}

func TestPrivateNetworkBlocklist_Addresses(t *testing.T) {
	tests := []struct {
		addr        string
		shouldBlock bool
	}{
		{"127.0.0.1:80", true},
		{"224.0.0.1:80", true},
		{"239.255.255.250:80", true},
		{"[::ffff:127.0.0.1]:80", true},
		{"[::ffff:10.0.0.1]:80", true},
		{"[::7f00:1]:80", true},
		{"[64:ff9b::7f00:1]:80", true},
		{"[64:ff9b::a9fe:a9fe]:80", true},
		{"[ff02::1]:80", true},
		{"[fe80::1]:80", true},
		{"[fd00::1]:80", true},
		{"[::a00:1]:80", true},
		{"93.184.216.34:80", false},
		{"[64:ff9b::5db8:d822]:80", false}, // NAT64 for 93.184.216.34
		{"[2606:2800:220:1::1]:80", false},
		{"[2001:4860:4802:32::a]:80", false},
	}
	for _, test := range tests {
		err := PrivateNetworkBlocklist.Control("tcp", test.addr, nil)
		var berr ErrBlocked
		if blocked := errors.As(err, &berr); blocked != test.shouldBlock {
			t.Errorf("%s: blocked = %v (err = %v), want %v", test.addr, blocked, err, test.shouldBlock)
		}
	}
}

func TestNetworkBlocklist_EmbeddedIPv4(t *testing.T) {
	blocklist := NetworkBlocklist{
		V4: []NamedNetwork{{IPNet: MustParseCIDR("10.0.0.0/8"), Name: "internal"}},
	}
	tests := []struct {
		addr        string
		shouldBlock bool
	}{
		{"10.0.0.1:80", true},
		{"[::a00:1]:80", true},
		{"[64:ff9b::a00:1]:80", true},
		{"[64:ff9b::b00:1]:80", false},
		{"[::1]:80", false},
		{"[2001:db8::a00:1]:80", false},
	}
	for _, test := range tests {
		err := blocklist.Control("tcp", test.addr, nil)
		var berr ErrBlocked
		if blocked := errors.As(err, &berr); blocked != test.shouldBlock {
			t.Errorf("%s: blocked = %v (err = %v), want %v", test.addr, blocked, err, test.shouldBlock)
		}
	}
}

// may be flaky
func TestNetworkBlocklist_Control(t *testing.T) {
	blocklist := NetworkBlocklist{
		V4: []NamedNetwork{
			{IPNet: MustParseCIDR("8.8.4.4/32")},
			{IPNet: MustParseCIDR("8.8.8.8/32")},
		},
		V6: []NamedNetwork{
			{IPNet: MustParseCIDR("2001:4860:4860::8844/128")},
			{IPNet: MustParseCIDR("2001:4860:4860::8888/128")},
		},
	}
	dialer := net.Dialer{
		Control: blocklist.Control,
	}

	ctx := context.Background()
	conn, err := dialer.DialContext(ctx, "udp", "8.8.8.8:53")
	assert.ErrorAs(t, err, &ErrBlocked{})
	if conn != nil {
		conn.Close()
	}

	conn, err = dialer.DialContext(ctx, "udp", "dns.google:53")
	assert.ErrorAs(t, err, &ErrBlocked{})
	if conn != nil {
		conn.Close()
	}
}

func ExampleNetworkBlocklist_Control() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   PrivateNetworkBlocklist.Control,
	}).DialContext

	client := &http.Client{
		Transport: transport,
	}

	_, err := client.Get("http://[::1]/")
	fmt.Println(err)
	// Output: Get "http://[::1]/": dial tcp [::1]:80: host is blocked (Loopback Address)
}
