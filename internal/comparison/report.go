// SPDX-License-Identifier: MIT
package comparison

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
)

func JSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// Guard CSV display cells from formula execution if opened in a spreadsheet.
// JSON retains exact input labels and is the lossless machine exchange format.
func csvText(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r\n", rune(s[0])) {
		return "'" + s
	}
	return s
}
func CSV(w io.Writer, r Report) error {
	c := csv.NewWriter(w)
	if err := c.Write([]string{"input", "source_sha256", "variant", "repetition", "status", "exit_code", "original_bytes", "candidate_bytes", "elapsed_ms", "width", "height", "transparent", "pixel_identical", "target_met", "candidate_sha256", "artifact", "error", "pipeline", "black_composite_rmse", "white_composite_rmse", "alpha_mae", "metric_samples", "metric_omission"}); err != nil {
		return err
	}
	for _, row := range r.Rows {
		pixel, target := "", ""
		if row.Metrics != nil {
			pixel = strconv.FormatBool(row.Metrics.PixelIdentical)
		}
		if row.TargetMet != nil {
			target = strconv.FormatBool(*row.TargetMet)
		}
		pipeline := []string{}
		for _, step := range row.Pipeline {
			pipeline = append(pipeline, step.Tool+"@"+step.Version)
		}
		black, white, alpha, samples, omission := "", "", "", "", ""
		score := func(v *float64) string {
			if v == nil {
				return ""
			}
			return strconv.FormatFloat(*v, 'g', -1, 64)
		}
		if row.Metrics != nil {
			black = score(row.Metrics.BlackRMSE)
			white = score(row.Metrics.WhiteRMSE)
			alpha = score(row.Metrics.AlphaMAE)
			samples = strconv.Itoa(row.Metrics.Samples)
			omission = row.Metrics.OmittedReason
		}
		if err := c.Write([]string{csvText(row.Input), row.SourceSHA256, csvText(row.Variant), strconv.Itoa(row.Repetition), row.Status, strconv.Itoa(row.ExitCode), strconv.Itoa(row.OriginalBytes), strconv.Itoa(row.CandidateBytes), strconv.FormatFloat(row.ElapsedMS, 'f', 3, 64), strconv.Itoa(row.After.Dimensions.Width), strconv.Itoa(row.After.Dimensions.Height), strconv.FormatBool(row.After.Transparency), pixel, target, row.CandidateSHA256, csvText(row.CandidatePath), csvText(row.Error), csvText(strings.Join(pipeline, " -> ")), black, white, alpha, samples, csvText(omission)}); err != nil {
			return err
		}
	}
	c.Flush()
	return c.Error()
}
func Human(w io.Writer, r Report) error {
	if _, err := fmt.Fprintln(w, "Local comparison: sizes are measurements, not a universal quality ranking."); err != nil {
		return err
	}
	for _, row := range r.Rows {
		if _, err := fmt.Fprintf(w, "%q | %s #%d | %d -> %d B | %.2f ms | %s (exit %d)\n", row.Input, row.Variant, row.Repetition, row.OriginalBytes, row.CandidateBytes, row.ElapsedMS, row.Status, row.ExitCode); err != nil {
			return err
		}
	}
	return nil
}
func HTML(r Report) ([]byte, error) {
	const source = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'"><title>smolpng local comparison</title><style>body{font:16px system-ui;max-width:1100px;margin:auto;padding:24px;color:#e6e9f1;background:#17202c}article{border-top:1px solid #657185;padding:20px 0}figure{display:inline-block;width:43%;vertical-align:top;margin:12px 2%}img{max-width:100%;max-height:300px;background:repeating-conic-gradient(#aaa 0% 25%,#fff 0% 50%) 50%/16px 16px}#matte-black:checked~article img{background:#111}#matte-white:checked~article img{background:#fff}label{margin-right:20px}input{margin-left:12px}code{overflow-wrap:anywhere}a{color:#9ee5cf}</style><h1>smolpng comparison</h1><p>Build {{.Version}} · <code>{{.Commit}}</code> · {{.Platform}}</p>{{range .Notes}}<p>{{.}}</p>{{end}}<p>No scripts, external resources, or live uploads. Preview backgrounds reveal transparent edges. Saved candidates are artifacts, not a declaration of perceptual quality.</p><p id="matte-caption">Preview background (keyboard-selectable):</p><input aria-describedby="matte-caption" type="radio" name="matte" id="matte-checker" checked><label for="matte-checker">Checkerboard</label><input type="radio" name="matte" id="matte-white"><label for="matte-white">White</label><input type="radio" name="matte" id="matte-black"><label for="matte-black">Black</label>{{range .Rows}}<article><h2>{{.Variant}} · run {{.Repetition}}</h2><p><code>{{.Input}}</code></p><p>{{.OriginalBytes}} → {{.CandidateBytes}} bytes · {{printf "%.2f" .ElapsedMS}} ms · {{.Status}} · exit {{.ExitCode}}</p>{{if .Error}}<p>{{.Error}}</p>{{end}}{{if .OriginalPath}}<figure><figcaption>Original snapshot</figcaption><img src="{{.OriginalPath}}" alt="Original source"><p><a href="{{.OriginalPath}}">Open original</a></p></figure>{{end}}{{if .CandidatePath}}<figure><figcaption>Candidate</figcaption><img src="{{.CandidatePath}}" alt="Compression candidate"><p><a href="{{.CandidatePath}}">Open candidate</a></p></figure>{{end}}{{if .Metrics}}<p>Decoded pixels identical: {{.Metrics.PixelIdentical}}. {{.Metrics.OmittedReason}}</p>{{end}}{{range .Warnings}}<p>Warning: {{.}}</p>{{end}}</article>{{end}}</html>`
	t, err := template.New("report").Parse(source)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = t.Execute(&b, r)
	return b.Bytes(), err
}
