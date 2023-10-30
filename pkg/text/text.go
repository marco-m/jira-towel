package text

import "fmt"

// ShortenMiddle returns 'input' shortened to 'width', by removing characters
// from the middle of the string. The returned string will have in the middle
// the filler "..".
func ShortenMiddle(input string, width int) string {
	const filler = ".."

	if width <= 0 || input == "" {
		return ""
	}
	if width >= len(input) {
		return input
	}
	if width < len(filler) {
		return filler[:width]
	}

	right := (width - len(filler)) / 2
	left := right
	if width%2 != 0 {
		left += 1
	}

	return fmt.Sprintf("%s%s%s", input[:left], filler, input[len(input)-right:])
}
