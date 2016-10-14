package apperix

import (
	"fmt"
	"bytes"
	"strings"
	"github.com/satori/go.uuid"
)

type Identifier struct {
	uuid uuid.UUID
}

func GenerateUniqueIdentifier() Identifier {
	return Identifier {
		uuid: uuid.NewV4(),
	}
}

func (identifier *Identifier) FromBytes(str []byte) {
	result, err := uuid.FromBytes(str)
	if err != nil {
		panic(fmt.Errorf(
			"Could not convert byte slice ('%s'(%d)) to unique identifier",
			str,
			len(str),
		))
	}
	identifier.uuid = result
}

func (identifier *Identifier) FromString(str string) {
	if len(str) != 32 {
		panic(fmt.Errorf(
			"Could not convert string ('%s') to unique identifier, string size is wrong (%d)",
			str,
			len(str),
		))
	}
	var buffer bytes.Buffer
	buffer.Write([]byte(str[0:8]))
	buffer.WriteRune('-')
	buffer.Write([]byte(str[8:12]))
	buffer.WriteRune('-')
	buffer.Write([]byte(str[12:16]))
	buffer.WriteRune('-')
	buffer.Write([]byte(str[16:20]))
	buffer.WriteRune('-')
	buffer.Write([]byte(str[20:]))
	result, err := uuid.FromString(buffer.String())
	if err != nil {
		panic(fmt.Errorf(
			"Could not convert string ('%s'(%d)) to unique identifier",
			str,
			len(str),
		))
	}
	identifier.uuid = result
}

func (identifier *Identifier) Bytes() []byte {
	return identifier.uuid.Bytes()
}

func (identifier *Identifier) String() string {
	str := identifier.uuid.String()
	str = strings.Replace(str, "-", "", -1)
	return str
}