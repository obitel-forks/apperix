package apperix

import (
	"bytes"
)

func ConcatStrings(base string, append ... string) string {
	var buffer bytes.Buffer
	buffer.WriteString(base)
	for _, item := range append {
		buffer.WriteString(item)
	}
	return buffer.String()
}