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
	"sync"
	"time"
)

type stat struct {
	mu    sync.Mutex
	lat   map[string][]time.Duration
	codes map[string]int
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
	flag.Parse()
	st := &stat{lat: map[string][]time.Duration{}, codes: map[string]int{}}
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
		return out, code
	}
	deadline := time.Now().Add(*dur)
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < *players; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			g, _ := call("guest", "POST", "/api/v1/auth/guest", "", map[string]any{}, false)
			d, _ := g["data"].(map[string]any)
			tok, _ := d["token"].(string)
			if tok == "" {
				return
			}
			call("join", "POST", fmt.Sprintf("/api/v1/worlds/%d/join", *world), tok, map[string]any{"name": fmt.Sprintf("Load %s", key()[:8]), "race": "ashborn"}, false)
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
				time.Sleep(200 * time.Millisecond)
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
	if bad > 0 {
		fmt.Println("FAIL: server/network errors present")
		os.Exit(1)
	}
}
