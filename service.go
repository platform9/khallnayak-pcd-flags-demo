// Demo "hamgr" service. Simulates a VMHA evacuation worker that consults
// a feature flag once per loop. Run it alongside flagd and flip the flag
// from a third terminal to watch behavior change in real-time.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	flagd "github.com/open-feature/go-sdk-contrib/providers/flagd/pkg"
	"github.com/open-feature/go-sdk/openfeature"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

func main() {
	host := envOr("FLAGD_HOST", "localhost")
	port, _ := strconv.Atoi(envOr("FLAGD_PORT", "8013"))

	fmt.Printf("%s== fake-hamgr starting; FLAGD=%s:%d ==%s\n",
		colorBold, host, port, colorReset)

	provider := flagd.NewProvider(
		flagd.WithHost(host),
		flagd.WithPort(uint16(port)),
	)
	openfeature.SetProvider(provider)
	client := openfeature.NewClient("hamgr")

	// Give the SDK a sec to subscribe to flagd's gRPC stream.
	time.Sleep(1500 * time.Millisecond)

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	cycle := 0
	for range tick.C {
		cycle++
		ctx := context.Background()
		fastEvac, _ := client.BooleanValue(ctx, "hamgr.fast_evac", false,
			openfeature.EvaluationContext{})

		ts := time.Now().Format("15:04:05")
		switch {
		case fastEvac:
			fmt.Printf("%s[%s] cycle=%d  %s🚀 FAST evac path: parallel-migrating 8 VMs in 12s%s\n",
				colorGreen, ts, cycle, colorBold, colorReset)
		default:
			fmt.Printf("%s[%s] cycle=%d   legacy evac path: serial-migrating 8 VMs in 96s%s\n",
				colorGray, ts, cycle, colorReset)
		}

		// One demo flair: alert if the operator just flipped the flag
		// — we keep a tiny in-process record to spot the transition.
		if cycle == 1 {
			lastVal = fastEvac
			continue
		}
		if fastEvac != lastVal {
			if fastEvac {
				fmt.Printf("%s            ↑ flag flipped ON — switched to fast path on cycle %d%s\n",
					colorYellow, cycle, colorReset)
			} else {
				fmt.Printf("%s            ↓ flag flipped OFF — rolled back to legacy path on cycle %d%s\n",
					colorYellow, cycle, colorReset)
			}
			lastVal = fastEvac
		}
	}
}

var lastVal bool

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
