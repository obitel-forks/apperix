package apperix

import (
	"bytes"
)

type resourceIdSegment struct {
	typ ResourceType
	identifier string
	name string
	value string
}

type ResourceIdentifier struct {
	path []resourceIdSegment
}

func (resId *ResourceIdentifier) Identifier() string {
	length := len(resId.path)
	if length < 1 {
		return "root"
	}
	return resId.path[length - 1].identifier
}

func (resId *ResourceIdentifier) VariableValues() map[string] string {
	result := make(map[string] string)
	var segment resourceIdSegment
	for _, segment = range resId.path {
		if segment.typ != VARIABLE {
			continue
		}
		result[segment.identifier] = segment.value
	}
	return result
}

func (resId *ResourceIdentifier) Url() string {
	var buffer bytes.Buffer
	var segment resourceIdSegment
	for _, segment = range resId.path {
		buffer.WriteRune('/')
		if segment.typ == VARIABLE {
			buffer.Write([]byte(segment.value))
		}
		buffer.Write([]byte(segment.name))
	}
	return buffer.String()
}

func (resId *ResourceIdentifier) String() string {
	var buffer bytes.Buffer
	var segment resourceIdSegment
	for _, segment = range resId.path {
		buffer.WriteRune('/')
		buffer.Write([]byte(segment.identifier))
		if segment.typ == VARIABLE {
			buffer.WriteRune('(')
			buffer.Write([]byte(segment.value))
			buffer.WriteRune(')')
		}
	}
	return buffer.String()
}

func (resId *ResourceIdentifier) Serialize() string {
	var buffer bytes.Buffer
	var segment resourceIdSegment
	if len(resId.path) < 1 {
		buffer.Write([]byte("root"))
		return buffer.String()
	}
	buffer.WriteString(resId.path[len(resId.path) - 1].identifier)
	for _, segment = range resId.path {
		if segment.typ != VARIABLE {
			continue
		}
		buffer.WriteRune('/')
		buffer.Write([]byte(segment.value))
	}
	return buffer.String()
}

func (resId *ResourceIdentifier) Parent() (
	parentId ResourceIdentifier,
	err error,
) {
	if len(resId.path) < 1 {
		return parentId, NotFoundError {
			message: "Root has no parent",
		}
	}
	parentPath := resId.path[:len(resId.path) - 1]
	return ResourceIdentifier {
		path: parentPath,
	}, nil
}

func (resId *ResourceIdentifier) HasParent() (bool) {
	return len(resId.path) > 0
}