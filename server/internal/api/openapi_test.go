package api

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestOpenAPICoversAllRoutes keeps api/openapi.yaml in sync with registered routes.
func TestOpenAPICoversAllRoutes(t *testing.T) {
	src, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(spec)
	re := regexp.MustCompile(`"(GET|POST|PUT|DELETE) (/[^"]+)"`)
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		method, path := strings.ToLower(m[1]), m[2]
		i := strings.Index(doc, "\n  "+path+":\n")
		if i < 0 {
			t.Errorf("route %s %s missing from openapi.yaml", m[1], path)
			continue
		}
		block := doc[i+1:]
		if j := strings.Index(block[3:], "\n  /"); j >= 0 {
			block = block[:j+3]
		}
		if !strings.Contains(block, "\n    "+method+":\n") {
			t.Errorf("method %s for %s missing from openapi.yaml", m[1], path)
		}
	}
}
