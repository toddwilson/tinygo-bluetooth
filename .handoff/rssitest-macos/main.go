// Test for Device.ReadRSSI. With no argument it lists nearby devices.
// With a name or address argument it connects and reads the RSSI.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

func main() {
	adapter := bluetooth.DefaultAdapter
	must(adapter.Enable())

	want := ""
	if len(os.Args) > 1 {
		want = strings.ToLower(os.Args[1])
	}

	seen := map[string]bluetooth.ScanResult{}
	var target *bluetooth.ScanResult
	time.AfterFunc(10*time.Second, func() { adapter.StopScan() })
	must(adapter.Scan(func(a *bluetooth.Adapter, r bluetooth.ScanResult) {
		seen[r.Address.String()] = r
		if want != "" && (strings.ToLower(r.Address.String()) == want ||
			(r.LocalName() != "" && strings.Contains(strings.ToLower(r.LocalName()), want))) {
			target = &r
			a.StopScan()
		}
	}))

	if target == nil {
		if want != "" {
			fmt.Println("no device matched", os.Args[1])
		}
		list := make([]bluetooth.ScanResult, 0, len(seen))
		for _, r := range seen {
			if r.LocalName() != "" {
				list = append(list, r)
			}
		}
		sort.Slice(list, func(i, j int) bool { return list[i].RSSI > list[j].RSSI })
		fmt.Printf("%d devices seen, %d with a name:\n", len(seen), len(list))
		for _, r := range list {
			fmt.Printf("  %4d dBm  %s  %q\n", r.RSSI, r.Address.String(), r.LocalName())
		}
		return
	}

	fmt.Printf("connecting to %q %s (advertised RSSI %d dBm)\n", target.LocalName(), target.Address.String(), target.RSSI)
	dev, err := adapter.Connect(target.Address, bluetooth.ConnectionParams{})
	must(err)
	defer dev.Disconnect()

	for i := 0; i < 5; i++ {
		start := time.Now()
		rssi, err := dev.ReadRSSI()
		fmt.Printf("ReadRSSI #%d: rssi=%d err=%v (%s)\n", i+1, rssi, err, time.Since(start).Round(time.Millisecond))
		time.Sleep(time.Second)
	}
}

func must(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
