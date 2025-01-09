package jira

import "fmt"

// CustomfieldValue returns the value of custom field 'name' from map
// 'customFields', which is assumed to be filled by github.com/mitchellh/mapstructure.
// See the tests in helpers_test for an example.
// If 'name' is not present, CustomfieldValue returns the empty string.
// CustomfieldValue assumes that lookup table 'lut' is filled by manual inspection
// of the JSON object returned by Jira.
// Yes, this sucks.
func CustomfieldValue(customFields map[string]any, lut map[string]int, name string) string {
	id, found := lut[name]
	if !found {
		return ""
	}
	cfName := fmt.Sprintf("customfield_%d", id)
	// A customfield JSON object has the following shape. We want the "value" field:
	// "customfield_11919": {
	//       "self": "https://x.atlassian.net/rest/api/2/customFieldOption/10837",
	//       "value": "Foo Bar", <=== THIS
	//       "id": "10837"
	//     },
	//
	// Convert from any to the expected shape, part 1
	cfMap, ok := customFields[cfName].(map[string]any)
	if !ok {
		return ""
	}
	// Convert from any to the expected shape, part 2
	value, ok := cfMap["value"].(string)
	if !ok {
		return ""
	}
	return value
}
