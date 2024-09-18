package text_test

import (
	"testing"

	"github.com/marco-m/rosina"

	"github.com/marco-m/jira-towel/pkg/text"
)

func TestTextShortenMiddle(t *testing.T) {
	type testCase struct {
		name  string
		input string
		width int
		want  string
	}

	testCases := []testCase{
		{
			name:  "empty, greater width",
			input: "",
			width: 10,
			want:  "",
		},
		{
			name:  "even, greater width",
			input: "1234567890",
			width: 20,
			want:  "1234567890",
		},
		{
			name:  "even, same width",
			input: "1234567890",
			width: 10,
			want:  "1234567890",
		},
		{
			name:  "even, smaller width 9",
			input: "1234567890",
			width: 9,
			want:  "1234..890",
		},
		{
			name:  "even, smaller width 8",
			input: "1234567890",
			width: 8,
			want:  "123..890",
		},

		{
			name:  "odd, greater width",
			input: "123456789",
			width: 20,
			want:  "123456789",
		},
		{
			name:  "odd, same width",
			input: "123456789",
			width: 9,
			want:  "123456789",
		},
		{
			name:  "odd, smaller width 8",
			input: "123456789",
			width: 8,
			want:  "123..789",
		},
		{
			name:  "odd, smaller width 7",
			input: "123456789",
			width: 7,
			want:  "123..89",
		},

		{
			name:  "width smaller than filling",
			input: "12345",
			width: 1,
			want:  ".",
		},
		{
			name:  "width same as filling",
			input: "12345",
			width: 2,
			want:  "..",
		},
		{
			name:  "0 width",
			input: "12345",
			width: 0,
			want:  "",
		},
		{
			name:  "negative width considered as 0",
			input: "12345",
			width: -3,
			want:  "",
		},
	}

	test := func(t *testing.T, tc testCase) {
		have := text.ShortenMiddle(tc.input, tc.width)

		rosina.AssertEqual(t, have, tc.want, "ShortenMiddle")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
