// SPDX-License-Identifier: MIT
package workflow

import (
	"context"
	"fmt"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/tinypng"
)

func (e Engine) Optimize(ctx context.Context, data []byte, r domain.Recipe) (Outcome, error) {
	out := Outcome{Warnings: []string{}}
	if err := r.Validate(); err != nil {
		return out, domain.E(2, err.Error(), err)
	}
	if err := ValidateExtras(r); err != nil {
		return out, domain.E(2, err.Error(), err)
	}
	if err := e.Limits.Validate(); err != nil {
		return out, domain.E(2, err.Error(), err)
	}
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}
	remoteOnly := r.Engine == "tinypng" || r.RemotePolicy == "always"
	localRecipe := r
	localRecipe.Engine = "local"
	localRecipe.RemotePolicy = "never"
	var localErr error
	if !remoteOnly {
		out, localErr = e.optimizeLocal(ctx, data, localRecipe)
	}
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	useRemote := remoteOnly || r.RemotePolicy == "compare" || r.RemotePolicy == "fallback" && (domain.Code(localErr) == 6 || domain.Code(localErr) == 4 || domain.Code(localErr) == 5)
	if !useRemote {
		return out, localErr
	}
	// Invalid input and preservation constraints never become reasons to upload.
	if localErr != nil && domain.Code(localErr) == 3 {
		return out, localErr
	}
	if r.Metadata == "keep" {
		return out, domain.E(2, "TinyPNG cannot guarantee keep-all metadata; choose safe/all explicitly before remote use", nil)
	}
	if r.Resize.Width > 0 || r.MaxDimensions.Width > 0 {
		return out, domain.E(2, "TinyPNG currently preserves original dimensions; explicit local resize is not combined with a remote upload", nil)
	}
	normalized, info, err := remoteInput(data, r, e.Limits.MaxPixels)
	if err != nil {
		return out, domain.E(3, err.Error(), err)
	}
	out.Original = info
	if e.Remote == nil {
		return out, domain.E(7, "TinyPNG adapter or credentials unavailable; use providers login tinypng", nil)
	}
	result, err := e.Remote.Optimize(ctx, tinypng.Request{Data: normalized, MaxBytes: e.Limits.MaxBytes})
	out.Remote = &result.Usage
	out.RemoteUsed = result.Usage.Attempted
	if err != nil {
		return out, err
	}
	candidate, err := remoteCandidate(normalized, result.Data, r, info, e.Limits.MaxPixels)
	if err != nil {
		return out, domain.E(7, "TinyPNG result rejected: "+err.Error(), err)
	}
	if remoteOnly {
		if !r.ForceOutput && len(candidate.Data) > len(data) {
			out.Candidate = Candidate{Data: normalized, Info: info, Rank: 100000000, Pipeline: []domain.Step{}}
			out.Warnings = append(out.Warnings, "TinyPNG result was larger; retained original pixels under the no-growth policy")
		} else {
			out.Candidate = candidate
			out.Remote.Selected = true
		}
	} else if (r.ForceOutput || len(candidate.Data) <= len(data)) && (len(out.Candidate.Data) == 0 || Better(candidate, &out.Candidate, r.TargetBytes)) {
		out.Candidate = candidate
		out.Remote.Selected = true
	}
	if localErr != nil && domain.Code(localErr) != 6 {
		out.Warnings = append(out.Warnings, "Local optimization unavailable; explicit remote policy supplied the candidate")
	}
	if len(out.Candidate.Data) == 0 {
		out.Candidate = Candidate{Data: normalized, Info: info, Rank: 100000000, Pipeline: []domain.Step{}}
		out.Warnings = append(out.Warnings, "Remote result was larger; retained original pixels")
	}
	if r.AllowResize {
		out.Warnings = append(out.Warnings, "TinyPNG used one original-dimension upload; remote resize searches are not performed to avoid multiplying quota use")
	}
	if r.TargetBytes > 0 && int64(len(out.Candidate.Data)) > r.TargetBytes {
		return out, domain.E(6, "target size was not reached; remote service supplies no enforceable quality/size target", nil)
	}
	return out, nil
}
func (e Engine) optimizeLocal(ctx context.Context, data []byte, r domain.Recipe) (Outcome, error) {
	return e.optimizePNG(ctx, data, r)
}
func remoteInput(data []byte, r domain.Recipe, maxPixels int64) ([]byte, domain.ImageInfo, error) {
	a, err := pngimage.Decode(data, maxPixels)
	if err != nil {
		return nil, domain.ImageInfo{}, err
	}
	if a.Info.Animated {
		return nil, a.Info, fmt.Errorf("animated input cannot be uploaded")
	}
	b, err := pngimage.Filter(data, r.Metadata)
	return b, a.Info, err
}
func remoteCandidate(source, data []byte, r domain.Recipe, before domain.ImageInfo, maxPixels int64) (Candidate, error) {
	a, err := pngimage.Decode(data, maxPixels)
	if err != nil {
		return Candidate{}, err
	}
	if a.Info.Dimensions != before.Dimensions || a.Info.Animated {
		return Candidate{}, fmt.Errorf("format, dimensions or animation changed unexpectedly")
	}
	if before.Transparency && !a.Info.Transparency {
		return Candidate{}, fmt.Errorf("transparency was discarded")
	}
	// Do not attach the input ICC to a remote service's potentially transformed pixels.
	data, err = pngimage.Reconcile(source, data, r.Metadata, true)
	if err != nil {
		return Candidate{}, err
	}
	return Candidate{Data: data, Info: a.Info, Rank: 0, Pipeline: []domain.Step{{Tool: "tinypng", Version: "HTTP API"}}}, nil
}
