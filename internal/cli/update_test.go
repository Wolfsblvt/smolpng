package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestUpdateStatusIsOfflineAndDefaultApplyRefusesNoTrust(t *testing.T) {
	var out, er bytes.Buffer
	env := Environment{Out: &out, Err: &er, In: strings.NewReader("")}
	if code := Run(context.Background(), []string{"update", "status", "--json"}, env); code != 0 || !json.Valid(out.Bytes()) {
		t.Fatal(code, out.String(), er.String())
	}
	out.Reset()
	if code := Run(context.Background(), []string{"update", "apply", "--json"}, env); code == 0 {
		t.Fatal("unconfigured update accepted")
	}
}
func TestAutomaticChecksNeedExplicitPolicyAndInterval(t *testing.T) {
	now := time.Now()
	for _, v := range []struct {
		p    string
		o, c bool
		t    time.Time
	}{{"manual", false, true, time.Time{}}, {"daily", true, true, time.Time{}}, {"daily", false, false, time.Time{}}, {"daily", false, true, now.Add(-time.Hour)}} {
		if updateCheckDue(v.p, v.o, v.c, v.t, now) {
			t.Fatal("unwanted automatic request", v)
		}
	}
	if !updateCheckDue("daily", false, true, now.Add(-25*time.Hour), now) {
		t.Fatal("opt-in interval blocked")
	}
}
