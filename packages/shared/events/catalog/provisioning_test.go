package catalog

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

// Compose files provision topics explicitly (producer auto-creation is disabled), so they must list exactly the
// catalog topics.
func TestComposeProvisionsExactlyTheCatalogTopics(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "infrastructure")
	provisioned := regexp.MustCompile(`hexta\.[a-z]+\.[a-z-]+\.v[0-9]+`)
	want := Topics()
	slices.Sort(want)
	for _, name := range []string{"compose.demo.yml", "compose.prod.yml"} {
		body, err := os.ReadFile(filepath.Join(root, name))
		if os.IsNotExist(err) {
			t.Skip("infrastructure not checked out with this module")
		}
		if err != nil {
			t.Fatal(err)
		}
		got := slices.Compact(slices.Sorted(slices.Values(provisioned.FindAllString(string(body), -1))))
		if !slices.Equal(got, want) {
			t.Errorf("%s provisions %v, catalog has %v", name, got, want)
		}
	}
}
