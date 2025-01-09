package jira_test

import (
	"encoding/json"
	"testing"

	"github.com/marco-m/jira-towel/pkg/jira"
	"github.com/marco-m/rosina"
	"github.com/mitchellh/mapstructure"
)

// The problem is that Jira, in the JSON response, mixes well-known fields
// (the fields declared by the API, always available) with numeric custom
// fields of the form "customfield_11919", AT THE SAME LEVEL. For example:
//
//	"fields": {
//	  "customfield_12011": ...
//	  "labels": [ ... ]
//	  "assignee": ...
//	}
//
// It would have been enough to simply put all the custom fields in a dedicated
// subobject, but unfortunately this is not how Jira does it:
//
//	"fields": {
//	  "labels": [ ... ]
//	  "assignee": ...
//	}
//	"customfields": {            <== WOULD HAVE BEEN SO EASY :-(
//	  "customfield_12011": ...
//	}
//
// So, we use the 'mapstructure' package to avoid as much as possible to copy
// fields by hand.
//
// 'mapstructure' looks for keys in JSON with the same name as the fields in
// the struct, but the comparison is case insensitive.
type Wrong struct {
	Reasonable int
	Normal     string
	// Tag 'remain' tells 'mapstructure' to collect all unknown fields, at any
	// level of nesting.
	Absurds map[string]any `mapstructure:",remain"`
}

func TestMapstructureCanHelp(t *testing.T) {
	inputJson := `{
  "reasonable":        42,
  "normal":            "i am normal",
  "customfield_1234":  {"value": "i am product X"},
  "customfield_3452":  {"value": "i am absurd 2"}
}`

	var parsedMap map[string]any
	err := json.Unmarshal([]byte(inputJson), &parsedMap)
	rosina.AssertNoError(t, err)

	// Only partial mapping, as happens in real Jira replies.
	lut := map[string]int{"product": 1234}
	var result Wrong
	err = mapstructure.Decode(parsedMap, &result)
	rosina.AssertNoError(t, err)

	have := jira.CustomfieldValue(result.Absurds, lut, "product")
	want := "i am product X"
	rosina.AssertEqual(t, have, want, "customfield existing")
}

func TestCustomFieldValue(t *testing.T) {
	lut := map[string]int{
		"product":  1234,
		"feature":  3452,
		"broken-1": 99,
		"broken-2": 34,
	}
	customfields := map[string]any{
		"customfield_1234": map[string]any{"value": "i am product X"},
		"customfield_3452": map[string]any{"value": "i am feature 2"},
		"customfield_99":   map[string]any{"broken": "i am broken"},
		"customfield_34":   "I am broken also",
	}

	have := jira.CustomfieldValue(customfields, lut, "product")
	want := "i am product X"
	rosina.AssertEqual(t, have, want, "customfield product")

	have = jira.CustomfieldValue(customfields, lut, "feature")
	want = "i am feature 2"
	rosina.AssertEqual(t, have, want, "customfield feature")

	have = jira.CustomfieldValue(customfields, lut, "non-existing")
	want = ""
	rosina.AssertEqual(t, have, want, "customfield non-existing")

	have = jira.CustomfieldValue(customfields, lut, "broken-1")
	want = ""
	rosina.AssertEqual(t, have, want, "customfield broken-1")

	have = jira.CustomfieldValue(customfields, lut, "broken-2")
	want = ""
	rosina.AssertEqual(t, have, want, "customfield broken-2")
}
