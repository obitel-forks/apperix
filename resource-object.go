package apperix

import (
	"fmt"
	"regexp"
)

type resourceObject interface {
	Identifier() string
	Name() string
	Handler(Method) (Handler, error)
	HasStaticChild(string) bool
	HasVariableChild(string) bool
	StaticChildIdentifier(string) (string, error)
	VariableChildIdentifier(int) (string, error)
	NumberOfVariableChildren() int
	DefineStaticChild(string, string)
	DefineVariableChild(string)
	Parent() string
	DefaultPermissions() DefaultResourcePermissions
}

type staticResource struct {
	identifier string
	name string
	parent string
	defaultPermissions DefaultResourcePermissions
	handlers map[Method] Handler
	staticChildren map[string] string
	variableChildren [] string
}

func (res *staticResource) Identifier() string {
	return res.identifier
}

func (res *staticResource) Name() string {
	return res.name
}

func (res *staticResource) Handler(method Method) (Handler, error) {
	if handler, exists := res.handlers[method]; exists {
		return handler, nil
	} else {
		return handler, fmt.Errorf("No handler found for method %s", method)
	}
}

func (res *staticResource) HasStaticChild(name string) bool {
	_, exists := res.staticChildren[name]
	return exists
}

func (res *staticResource) HasVariableChild(name string) bool {
	for _, value := range res.variableChildren {
		if value == name {
			return true
		}
	}
	return false
}

func (res *staticResource) StaticChildIdentifier(name string) (string, error) {
	result, exists := res.staticChildren[name]
	if !exists {
		return "", fmt.Errorf("Static child (%s) out of range", name)
	}
	return result, nil
}

func (res *staticResource) VariableChildIdentifier(index int) (string, error) {
	if index >= len(res.variableChildren) {
		return "", fmt.Errorf("Variable child index (%d) out of range", index)
	}
	return res.variableChildren[index], nil
}

func (res *staticResource) NumberOfVariableChildren() int {
	return len(res.variableChildren)
}

func (res *staticResource) DefineStaticChild(identifier string, name string) {
	res.staticChildren[name] = identifier
}

func (res *staticResource) DefineVariableChild(identifier string) {
	res.variableChildren = append(res.variableChildren, identifier)
}

func (res *staticResource) Parent() string {
	return res.parent
}

func (res *staticResource) DefaultPermissions() DefaultResourcePermissions {
	return res.defaultPermissions
}

type variableResource struct {
	identifier string
	name string
	parent string
	defaultPermissions DefaultResourcePermissions
	handlers map[Method] Handler
	staticChildren map[string] string
	variableChildren [] string
	pattern regexp.Regexp
}

func (res *variableResource) Identifier() string {
	return res.identifier
}

func (res *variableResource) Name() string {
	return res.name
}

func (res *variableResource) Handler(method Method) (Handler, error) {
	if handler, exists := res.handlers[method]; exists {
		return handler, nil
	} else {
		return handler, fmt.Errorf("No handler found for method %s", method)
	}
}

func (res *variableResource) HasStaticChild(name string) bool {
	_, exists := res.staticChildren[name]
	return exists
}

func (res *variableResource) HasVariableChild(name string) bool {
	for _, value := range res.variableChildren {
		if value == name {
			return true
		}
	}
	return false
}

func (res *variableResource) DefineStaticChild(identifier string, name string) {
	res.staticChildren[name] = identifier
}

func (res *variableResource) DefineVariableChild(identifier string) {
	res.variableChildren = append(res.variableChildren, identifier)
}

func (res *variableResource) StaticChildIdentifier(name string) (string, error) {
	result, exists := res.staticChildren[name]
	if !exists {
		return "", fmt.Errorf("Static child (%s) out of range", name)
	}
	return result, nil
}

func (res *variableResource) VariableChildIdentifier(index int) (string, error) {
	if index >= len(res.variableChildren) {
		return "", fmt.Errorf("Variable child index (%d) out of range", index)
	}
	return res.variableChildren[index], nil
}

func (res *variableResource) NumberOfVariableChildren() int {
	return len(res.variableChildren)
}

func (res *variableResource) MatchPattern(str string) bool {
	return res.pattern.MatchString(str)
}

func (res *variableResource) Parent() string {
	return res.parent
}

func (res *variableResource) DefaultPermissions() DefaultResourcePermissions {
	return res.defaultPermissions
}
