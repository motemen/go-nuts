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
		{"[::2]", false},
		{"[2001:2::1]", true},
		{"[2001:4860:4802:32::a]", false},
		{"[::ffff:192.168.0.1]", true},
	}

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		req, _ := http.NewRequestWithContext(ctx, "GET", "http://"+test.addr, nil)
		_, err := client.Do(req)
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
