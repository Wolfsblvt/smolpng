package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The JSON schema is a distributed automation artifact, not a separate model of
// the compressor. Keep its property/required field declarations aligned with
// the actual metadata types, including omitempty and excluded image bytes.
func TestCompressionSchemaTracksActualJSONFields(t *testing.T) {
	b, e := os.ReadFile(filepath.Join("..", "..", "schemas", "compression-v1.schema.json"))
	if e != nil {
		t.Fatal(e)
	}
	var schema struct {
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		} `json:"$defs"`
	}
	if e = json.Unmarshal(b, &schema); e != nil {
		t.Fatal(e)
	}
	for name, value := range map[string]any{"dimensions": Dimensions{}, "dimensionChange": DimensionChange{}, "step": Step{}, "target": Target{}, "result": Result{}, "summary": Summary{}, "report": Report{}, "remoteUsage": RemoteUsage{}, "jpegAttachment": JPEGAttachment{}} {
		t.Run(name, func(t *testing.T) {
			def, ok := schema.Definitions[name]
			if !ok {
				t.Fatal("schema definition missing")
			}
			properties := []string{}
			required := []string{}
			typ := reflect.TypeOf(value)
			for i := 0; i < typ.NumField(); i++ {
				tag := typ.Field(i).Tag.Get("json")
				p := strings.Split(tag, ",")
				if p[0] == "-" {
					continue
				}
				properties = append(properties, p[0])
				if !strings.Contains(tag, ",omitempty") {
					required = append(required, p[0])
				}
			}
			actual := []string{}
			for key := range def.Properties {
				actual = append(actual, key)
			}
			sort.Strings(actual)
			sort.Strings(properties)
			sort.Strings(required)
			sort.Strings(def.Required)
			if !reflect.DeepEqual(actual, properties) || !reflect.DeepEqual(required, def.Required) {
				t.Fatal("schema/JSON tags diverged", actual, properties, def.Required, required)
			}
		})
	}
	r := NewResult("sample", "auto")
	r.Bytes = []byte("private image bytes")
	encoded, e := json.Marshal(Summarize([]Result{r}))
	if e != nil || strings.Contains(string(encoded), "private image bytes") || strings.Contains(string(encoded), "cHJpdmF0ZSBpbWFnZSBieXRlcw") {
		t.Fatal("binary image leaked into metadata", e)
	}
}
