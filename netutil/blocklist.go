// Package netutil provides net-related utility functions/types.
package netutil

import (
	"fmt"
	"net"
	"syscall"
)

// PrivateNetworkBlocklist is a blocklist that blocks dialing to private networks.
var PrivateNetworkBlocklist NetworkBlocklist

// NetworkBlocklist is a blocklist that blocks dialing to specified networks.
type NetworkBlocklist struct {
	V4 []NamedNetwork
	V6 []NamedNetwork
}

// Control is intended to be passed to net.Dialer.Control in order to block dialing to networks specified in l.
func (l NetworkBlocklist) Control(network, address string, c syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("cannot parse address %q: %w", address, err)
	}

	addr := net.ParseIP(host)
	if addr == nil {
		return fmt.Errorf("cannot parse host %q", host)
	}

	if ip4 := addr.To4(); ip4 != nil {
		if n, ok := lookup(l.V4, ip4); ok {
			return ErrBlocked{Host: host, Network: n}
		}
	} else if addr.To16() != nil {
		if n, ok := lookup(l.V6, addr); ok {
			return ErrBlocked{Host: host, Network: n}
		}
		// Addresses that embed an IPv4 address (NAT64, IPv4-compatible) may reach that
		// IPv4 address, so check the embedded address against the IPv4 list too.
		if ip4 := embeddedIPv4(addr); ip4 != nil {
			if n, ok := lookup(l.V4, ip4); ok {
				return ErrBlocked{Host: host, Network: n}
			}
		}
	} else {
		return fmt.Errorf("BUG: unreachable")
	}

	return nil
}

func lookup(networks []NamedNetwork, ip net.IP) (NamedNetwork, bool) {
	for _, n := range networks {
		if n.IPNet.Contains(ip) {
			return n, true
		}
	}
	return NamedNetwork{}, false
}

var (
	// NAT64 well-known prefix (RFC 6052). Where NAT64 is available, 64:ff9b::7f00:1 may reach 127.0.0.1.
	nat64WellKnownPrefix = MustParseCIDR("64:ff9b::/96")
	// Deprecated IPv4-compatible addresses (RFC 4291). net.IP.To4 does not treat them as IPv4.
	ipv4CompatiblePrefix = MustParseCIDR("::/96")
)

// embeddedIPv4 returns the IPv4 address embedded in a NAT64 well-known prefix address
// or an IPv4-compatible address, or nil if ip is neither.
func embeddedIPv4(ip net.IP) net.IP {
	ip = ip.To16()
	if ip == nil {
		return nil
	}
	if nat64WellKnownPrefix.Contains(ip) || ipv4CompatiblePrefix.Contains(ip) {
		return net.IPv4(ip[12], ip[13], ip[14], ip[15]).To4()
	}
	return nil
}

// ErrBlocked is an error returned by NetworkBlocklist.Control (thus net.Dialer.DialContext) when
// outgoing host is blocked by NetworkBlocklist.
type ErrBlocked struct {
	Host    string
	Network NamedNetwork
}

func (e ErrBlocked) Error() string {
	message := "host is blocked"
	if e.Network.Name != "" {
		message += fmt.Sprintf(" (%s)", e.Network.Name)
	}
	return message
}

type NamedNetwork struct {
	IPNet *net.IPNet
	Name  string
}

type unparsedNamedNetwork struct {
	ip   string
	name string
}

func MustParseCIDR(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Sprintf("cannot parse CIDR %q: %s", cidr, err))
	}

	return ipNet
}

func init() {
	PrivateNetworkBlocklist.V4 = make([]NamedNetwork, len(privateNetworksV4))
	for i, p := range privateNetworksV4 {
		PrivateNetworkBlocklist.V4[i] = NamedNetwork{
			IPNet: MustParseCIDR(p.ip),
			Name:  p.name,
		}
	}

	PrivateNetworkBlocklist.V6 = make([]NamedNetwork, len(privateNetworksV6))
	for i, p := range privateNetworksV6 {
		PrivateNetworkBlocklist.V6[i] = NamedNetwork{
			IPNet: MustParseCIDR(p.ip),
			Name:  p.name,
		}
	}
}

// https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
var privateNetworksV4 = []unparsedNamedNetwork{
	{"0.0.0.0/8", `"This network"`},
	{"0.0.0.0/32", `"This host on this network"`},
	{"127.0.0.0/8", "Loopback"},
	{"255.255.255.255/32", "Limited Broadcast"},
	{"240.0.0.0/4", "Reserved"},
	{"10.0.0.0/8", "Private-Use"},
	{"172.16.0.0/12", "Private-Use"},
	{"192.168.0.0/16", "Private-Use"},
	{"198.18.0.0/15", "Benchmarking"},
	{"192.88.99.0/24", "Deprecated (6to4 Relay Anycast)"},
	{"169.254.0.0/16", "Link Local"},
	{"192.0.0.0/24", "IETF Protocol Assignments"},
	{"192.0.2.0/24", "Documentation (TEST-NET-1)"},
	{"198.51.100.0/24", "Documentation (TEST-NET-2)"},
	{"203.0.113.0/24", "Documentation (TEST-NET-3)"},
	{"192.0.0.0/29", "IPv4 Service Continuity Prefix"},
	{"100.64.0.0/10", "Shared Address Space"},
	{"192.0.0.170/32", "NAT64/DNS64 Discovery"},
	{"192.0.0.171/32", "NAT64/DNS64 Discovery"},
	{"192.0.0.8/32", "IPv4 dummy address"},
	{"224.0.0.0/4", "Multicast"},
}

// https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
var privateNetworksV6 = []unparsedNamedNetwork{
	{"2001::/23", "IETF Protocol Assignments"},
	{"2002::/16", "6to4"},
	{"2001:db8::/32", "Documentation"},
	{"fc00::/7", "Unique-Local"},
	{"2001::/32", "TEREDO"},
	{"::1/128", "Loopback Address"},
	{"::/128", "Unspecified Address"},
	{"::ffff:0:0/96", "IPv4-mapped Address"},
	{"fe80::/10", "Link-Local Unicast"},
	{"2001:10::/28", "Deprecated (previously ORCHID)"},
	{"2001:2::/48", "Benchmarking"},
	{"100::/64", "Discard-Only Address Block"},
	{"64:ff9b:1::/48", "IPv4-IPv6 Translat."},
	{"ff00::/8", "Multicast"},
}
