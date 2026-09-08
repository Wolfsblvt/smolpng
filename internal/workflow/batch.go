// SPDX-License-Identifier: MIT
package workflow

import (
	"bytes"
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type Batch struct {
	Engine   Engine
	Resolver input.Resolver
	Stdout   io.Writer
}
type Options struct {
	Recipe                    domain.Recipe
	Output                    output.Options
	Jobs                      int
	FailFast, DryRun, Offline bool
}

func (b Batch) Run(ctx context.Context, jobs []output.Job, o Options) domain.Report {
	results := make([]domain.Result, len(jobs))
	workers := max(1, min(o.Jobs, 32))
	if o.FailFast {
		workers = 1
	}
	queue := make(chan int)
	var wg sync.WaitGroup
	var failed atomic.Bool
	for range workers {
		wg.Go(func() {
			for i := range queue {
				r := domain.NewResult(jobs[i].Source.Display, o.Recipe.Name)
				if ctx.Err() != nil {
					r.Fail(ctx.Err())
				} else if o.FailFast && failed.Load() {
					r.Status = "skipped"
					r.Warnings = append(r.Warnings, "not started after previous failure")
				} else {
					r = b.one(ctx, jobs[i], o)
				}
				if r.ExitCode != 0 {
					failed.Store(true)
				}
				results[i] = r
			}
		})
	}
	for i := range jobs {
		queue <- i
	}
	close(queue)
	wg.Wait()
	return domain.Summarize(results)
}
func (b Batch) one(ctx context.Context, j output.Job, o Options) domain.Result {
	r := domain.NewResult(j.Source.Display, o.Recipe.Name)
	if j.Source.Err != nil {
		r.Fail(j.Source.Err)
		return r
	}
	if o.DryRun {
		r.Status = "planned"
		r.Output = j.Output
		return r
	}
	if o.Recipe.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.Recipe.Timeout)
		defer cancel()
	}
	data, er := b.Resolver.Load(ctx, j.Source, b.Engine.Limits.MaxBytes, o.Offline)
	if er != nil {
		r.Fail(er)
		return r
	}
	out, er := b.Engine.Optimize(ctx, data, o.Recipe)
	r.Remote = out.Remote
	r.RemoteUsed = out.RemoteUsed
	r.OriginalBytes = int64(len(data))
	r.Format = out.Original.Format
	r.Dimensions.Before = out.Original.Dimensions
	r.Warnings = append(r.Warnings, out.Warnings...)
	if er != nil && domain.Code(er) != 6 {
		r.Fail(er)
		return r
	}
	c := out.Candidate
	r.OutputBytes = int64(len(c.Data))
	r.SavedBytes = r.OriginalBytes - r.OutputBytes
	if r.OriginalBytes > 0 {
		r.SavingPercent = 100 * float64(r.SavedBytes) / float64(r.OriginalBytes)
	}
	r.Dimensions.After = c.Info.Dimensions
	r.ResizeUsed = r.Dimensions.Before != r.Dimensions.After
	r.Pipeline = c.Pipeline
	if r.Pipeline == nil {
		r.Pipeline = []domain.Step{}
	}
	if o.Recipe.TargetBytes > 0 {
		r.Target = &domain.Target{RequestedBytes: o.Recipe.TargetBytes, Met: r.OutputBytes <= o.Recipe.TargetBytes}
		if !r.Target.Met {
			r.Target.NearestBytes = r.OutputBytes
		}
	}
	if er != nil && !o.Recipe.AcceptNearest {
		r.Fail(er)
		return r
	}
	if o.Output.InPlace {
		current, e := os.ReadFile(j.Source.Path)
		if e != nil || !bytes.Equal(current, data) {
			r.Fail(domain.E(5, "original changed during compression; refusing in-place replacement", e))
			return r
		}
	}
	var writeErr error
	if j.Output == "-" {
		if b.Stdout == nil {
			writeErr = domain.E(5, "binary stdout unavailable", nil)
		} else {
			_, writeErr = b.Stdout.Write(c.Data)
		}
	} else {
		writeErr = output.Write(ctx, j, c.Data, o.Output)
	}
	if writeErr != nil {
		r.Fail(domain.E(5, "cannot commit complete output", writeErr))
		return r
	}
	r.Output = j.Output
	r.Status = "optimized"
	if bytes.Equal(data, c.Data) {
		r.Status = "unchanged"
	}
	if er != nil {
		r.Fail(er)
	}
	return r
}
