package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

func waitforNetwork() error {

	maxWait := time.Second * 3
	checkinterval := time.Second
	timestarted := time.Now()

	for {

		interfaces, err := net.Interfaces()
		if err != nil {

			return err
		}

		if len(interfaces) > 1 {

			return nil
		}

		if time.Since(timestarted) > maxWait {

			return fmt.Errorf("timeout after  %s", maxWait)
		}

		time.Sleep(checkinterval)

	}
}

func setupNAT(bridgeName string, subnet string) error {

	err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)

	if err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	ipt, err := iptables.New()

	if err != nil {
		return fmt.Errorf("failed to get iptables handle: %w", err)
	}

	err = ipt.AppendUnique("nat", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE")
	if err != nil {
		return fmt.Errorf("failed to add MASQUERADE rule: %w", err)
	}

	fmt.Println(" NAT rules set up FINALLLY!!!!!!!!!!!!!!.")
	return nil

}

func cleanupNAT(bridgeName string, subnet string) {

	ipt, err := iptables.New()

	if err != nil {
		fmt.Printf("Warning: failed to get iptables handle for cleanup: %v\n", err)
		return
	}

	err = ipt.Delete("nat", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE")
	if err != nil {
		fmt.Printf(" MASQUERADE rule  may not exist: %v\n", err)
	}
}

func setupDNS(newRoot string) error {

	resolvConfPath := filepath.Join(newRoot, "etc", "resolv.conf")

	os.MkdirAll(filepath.Dir(resolvConfPath), 0755)

	dnsConfig := "nameserver 8.8.8.8\n"

	if err := os.WriteFile(resolvConfPath, []byte(dnsConfig), 0644); err != nil {
		return fmt.Errorf("failed to write resolv.conf in new root: %w", err)
	}

	return nil
}
