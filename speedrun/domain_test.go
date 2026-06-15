package speedrun

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring. The client HTTP behaviour is covered in speedrun_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "speedrun" {
		t.Errorf("Scheme = %q, want speedrun", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "speedrun" {
		t.Errorf("Identity.Binary = %q, want speedrun", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("pd0wq31e")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if typ != "game" {
		t.Errorf("type = %q, want game", typ)
	}
	if id != "pd0wq31e" {
		t.Errorf("id = %q, want pd0wq31e", id)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("game", "sm64")
	want := "https://" + Host + "/sm64"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestHostWiring(t *testing.T) {
	_, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
}
