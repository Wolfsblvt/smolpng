package workflow

import (
	"bytes"
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/Wolfsblvt/smolpng/internal/tinypng"
	"image/png"
	"testing"
)

type remoteStub struct {
	calls int
	data  []byte
	input []byte
}

func (s *remoteStub) Optimize(_ context.Context, r tinypng.Request) (tinypng.Response, error) {
	s.calls++
	s.input = bytes.Clone(r.Data)
	return tinypng.Response{Data: s.data, Usage: domain.RemoteUsage{Provider: "tinypng", Attempted: true}}, nil
}
func TestLocalPoliciesNeverUpload(t *testing.T) {
	data := testutil.PNG(testutil.Pattern(16, 16, "flat"), png.NoCompression)
	for _, policy := range []string{"never", "ask", "fallback"} {
		stub := &remoteStub{}
		e := Engine{Limits: domain.DefaultLimits(), Remote: stub}
		r, _ := domain.Builtin("auto")
		r.TargetBytes = 1 << 20
		r.RemotePolicy = policy
		_, err := e.Optimize(context.Background(), data, r)
		if err != nil || stub.calls != 0 {
			t.Fatal(policy, err, stub.calls)
		}
	}
}
func TestRemoteReportedEvenWhenOriginalWins(t *testing.T) {
	data := testutil.PNG(testutil.Pattern(16, 16, "flat"), png.BestCompression)
	b := testutil.PNG(testutil.Pattern(16, 16, "flat"), png.NoCompression)
	s := &remoteStub{data: b}
	e := Engine{Limits: domain.DefaultLimits(), Remote: s}
	r, _ := domain.Builtin("auto")
	r.RemotePolicy = "always"
	out, err := e.Optimize(context.Background(), data, r)
	if err != nil || !out.RemoteUsed || out.Remote.Selected || !bytes.Equal(out.Candidate.Data, data) || s.calls != 1 {
		t.Fatal(out, err)
	}
}
func TestRemoteCannotBypassLosslessOrKeep(t *testing.T) {
	data := testutil.PNG(testutil.Pattern(16, 16, "flat"), png.NoCompression)
	s := &remoteStub{data: data}
	e := Engine{Limits: domain.DefaultLimits(), Remote: s}
	r, _ := domain.Builtin("lossless")
	r.RemotePolicy = "always"
	if _, er := e.Optimize(context.Background(), data, r); domain.Code(er) != 2 {
		t.Fatal(er)
	}
	r, _ = domain.Builtin("auto")
	r.Metadata = "keep"
	r.RemotePolicy = "always"
	if _, er := e.Optimize(context.Background(), data, r); domain.Code(er) != 2 || s.calls != 0 {
		t.Fatal(er)
	}
}
