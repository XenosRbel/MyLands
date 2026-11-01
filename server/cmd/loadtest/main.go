// Command loadtest drives N concurrent players through the real HTTP API.
//
//	go run ./cmd/loadtest -url http://127.0.0.1:8080 -world 1 -players 200 -duration 30s
//
// Each virtual player: guest sign-in → join → loop {city view | build | train}
// with idempotency keys. Reports throughput, latency percentiles and errors by code.
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type stat struct {
	mu    sync.Mutex
	lat   map[string][]time.Duration
	codes map[string]int
	cmdAt []time.Duration // offsets of build and train calls from the test start
}

func (s *stat) add(op string, d time.Duration, code string) {
	s.mu.Lock()
	s.lat[op] = append(s.lat[op], d)
	s.codes[code]++
	s.mu.Unlock()
}

func key() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	url := flag.String("url", "http://127.0.0.1:8080", "server base URL")
	world := flag.Int("world", 1, "world id")
	players := flag.Int("players", 100, "virtual players")
	dur := flag.Duration("duration", 30*time.Second, "test duration")
	think := flag.Duration("think", 200*time.Millisecond, "pause between one player's actions (raise it to lower the command rate)")
	ramp := flag.Duration("ramp", 0, "spread player start times over this period (0 = join burst)")
	race := flag.String("race", "ashborn", "race id used when joining")
	metricsToken := flag.String("metrics-token", "", "bearer token for /metrics (empty if none is set)")
	targetP95 := flag.Duration("target-p95", 0, "fail if p95 of city, build or train exceeds this (0 = no check)")
	targetCmds := flag.Float64("target-cmds", 0, "fail if commands per second fall below this (0 = no check)")
	targetLag := flag.Float64("target-lag", 0, "fail if worker lag at the end exceeds this many seconds (0 = no check)")
	flag.Parse()
	st := &stat{lat: map[string][]time.Duration{}, codes: map[string]int{}}
	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{MaxIdleConnsPerHost: 512}}
	call := func(op, method, path, tok string, body any, idem bool) (map[string]any, string) {
		var r io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			r = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, *url+path, r)
		req.Header.Set("Content-Type", "application/json")
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		if idem {
			req.Header.Set("Idempotency-Key", key())
		}
		t0 := time.Now()
		res, err := client.Do(req)
		if err != nil {
			st.add(op, time.Since(t0), "neterr")
			return nil, "neterr"
		}
		defer res.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(res.Body).Decode(&out)
		code := fmt.Sprint(res.StatusCode)
		if e, ok := out["error"].(map[string]any); ok {
			code += ":" + fmt.Sprint(e["code"])
		}
		st.add(op, time.Since(t0), code)
		if op == "build" || op == "train" {
			st.mu.Lock()
			st.cmdAt = append(st.cmdAt, time.Since(start))
			st.mu.Unlock()
		}
		return out, code
	}
	deadline := start.Add(*dur)
	var wg sync.WaitGroup
	for i := 0; i < *players; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if *ramp > 0 {
				time.Sleep(*ramp * time.Duration(i) / time.Duration(*players))
			}
			g, _ := call("guest", "POST", "/api/v1/auth/guest", "", map[string]any{}, false)
			d, _ := g["data"].(map[string]any)
			tok, _ := d["token"].(string)
			if tok == "" {
				return
			}
			call("join", "POST", fmt.Sprintf("/api/v1/worlds/%d/join", *world), tok, map[string]any{"name": fmt.Sprintf("Load %s", key()[:8]), "race": *race}, false)
			ov, _ := call("overview", "GET", fmt.Sprintf("/api/v1/worlds/%d/overview", *world), tok, nil, false)
			od, _ := ov["data"].(map[string]any)
			cities, _ := od["cities"].([]any)
			if len(cities) == 0 {
				return
			}
			cid := int(cities[0].(map[string]any)["id"].(float64))
			blds := []string{"lumberyard", "farmstead", "quarry", "cottages", "warehouse", "keep"}
			for n := 0; time.Now().Before(deadline); n++ {
				switch n % 4 {
				case 0, 2:
					call("city", "GET", fmt.Sprintf("/api/v1/cities/%d", cid), tok, nil, false)
				case 1:
					call("build", "POST", fmt.Sprintf("/api/v1/cities/%d/build", cid), tok, map[string]any{"building": blds[(n+i)%len(blds)]}, true)
				case 3:
					call("train", "POST", fmt.Sprintf("/api/v1/cities/%d/train", cid), tok, map[string]any{"unit": "militia", "quantity": 1}, true)
				}
				time.Sleep(*think)
			}
		}(i)
	}
	wg.Wait()
	el := time.Since(start)
	total := 0
	ops := make([]string, 0, len(st.lat))
	for op := range st.lat {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	fmt.Printf("players=%d duration=%s\n", *players, el.Round(time.Millisecond))
	fmt.Printf("%-9s %7s %8s %8s %8s %8s\n", "op", "count", "p50", "p95", "p99", "max")
	for _, op := range ops {
		l := st.lat[op]
		sort.Slice(l, func(a, b int) bool { return l[a] < l[b] })
		total += len(l)
		pct := func(p float64) time.Duration { return l[min(len(l)-1, int(float64(len(l))*p))] }
		fmt.Printf("%-9s %7d %8s %8s %8s %8s\n", op, len(l), pct(.5).Round(time.Microsecond*100), pct(.95).Round(time.Microsecond*100), pct(.99).Round(time.Microsecond*100), l[len(l)-1].Round(time.Microsecond*100))
	}
	fmt.Printf("throughput: %.1f req/s\nstatus codes:\n", float64(total)/el.Seconds())
	codes := make([]string, 0, len(st.codes))
	for c := range st.codes {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	bad := 0
	for _, c := range codes {
		fmt.Printf("  %-40s %d\n", c, st.codes[c])
		if c == "neterr" || c[0] == '5' {
			bad += st.codes[c]
		}
	}
	failed := bad > 0
	if bad > 0 {
		fmt.Println("FAIL: server/network errors present")
	}
	cmds := len(st.lat["build"]) + len(st.lat["train"])
	// The rate is measured after the ramp, when every player is in the game.
	steady := 0
	for _, at := range st.cmdAt {
		if at >= *ramp+5*time.Second {
			steady++
		}
	}
	window := el - *ramp - 5*time.Second
	rate := float64(cmds) / el.Seconds()
	if window > 0 {
		rate = float64(steady) / window.Seconds()
	}
	fmt.Printf("commands: %d total, %.1f/s in the steady window (after ramp + 5s)\n", cmds, rate)
	if *targetCmds > 0 && rate < *targetCmds {
		fmt.Printf("FAIL: %.1f commands/s is below target %.1f\n", rate, *targetCmds)
		failed = true
	}
	if *targetP95 > 0 {
		for _, op := range []string{"city", "build", "train"} {
			l := st.lat[op]
			if len(l) == 0 {
				continue
			}
			if p := l[min(len(l)-1, int(float64(len(l))*.95))]; p > *targetP95 {
				fmt.Printf("FAIL: %s p95 %s exceeds target %s\n", op, p.Round(time.Microsecond*100), *targetP95)
				failed = true
			}
		}
	}
	m, err := scrape(client, *url, *metricsToken)
	if err != nil {
		fmt.Println("metrics: unavailable:", err)
	} else {
		fmt.Printf("metrics: events processed %.0f, failed %.0f, overdue %.0f, parked %.0f, max lag %.1fs, tx retries %.0f, db conns %.0f/%.0f\n",
			m["rf_events_processed_total"], m["rf_events_failed_total"], m["rf_events_overdue"], m["rf_events_parked"],
			m["rf_events_max_lag_seconds"], m["rf_tx_retries_total"], m["rf_db_conns_acquired"], m["rf_db_conns_total"])
		if *targetLag > 0 && m["rf_events_max_lag_seconds"] > *targetLag {
			fmt.Printf("FAIL: worker lag %.1fs exceeds target %.1fs\n", m["rf_events_max_lag_seconds"], *targetLag)
			failed = true
		}
		if m["rf_events_failed_total"] > 0 || m["rf_events_parked"] > 0 {
			fmt.Println("FAIL: failed or parked events")
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

// scrape reads the plain-text /metrics endpoint into name -> value.
func scrape(c *http.Client, base, token string) (map[string]float64, error) {
	req, _ := http.NewRequest("GET", base+"/metrics", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}
	out := map[string]float64{}
	data, _ := io.ReadAll(res.Body)
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 {
			if v, err := strconv.ParseFloat(f[1], 64); err == nil {
				out[f[0]] = v
			}
		}
	}
	return out, nil
}
