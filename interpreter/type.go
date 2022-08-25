package main

import (
	"fmt"
	"strings"
)

type Type interface {
	// Equals returns true if both types are equal
	Equals(other Type) bool
	// ZeroValue returns the default value to be returned when a
	// variable of this type is declared without explicit initialization
	ZeroValue() Value
	fmt.Stringer
}

// newAtomicType returns a new atomic type based on an underlying Go type.
func newAtomicType[UnderlyingGoType Value](name string) Type {
	return &atomicType[UnderlyingGoType]{name: name}
}

// atomicType represents one of the four atomic types
// in B-minor: "string", "boolean", "char", and "integer".
type atomicType[UnderlyingGoType Value] struct {
	name      string
	zeroValue UnderlyingGoType
}

func (t *atomicType[UnderlyingGoType]) String() string {
	return t.name
}

func (t *atomicType[UnderlyingGoType]) Equals(other Type) bool {
	return t == other
}

// ZeroValue returns the default value to be assigned to
// B-minor variables that are declared without an initializer.
// For example:
//
//	x: integer; // x has the zero integer value, 0
//	y: boolean; // y has the zero boolean value, false
//
// The zero value of an atomic type is the zero value of
// its underlying Go type.
func (t *atomicType[UnderlyingGoType]) ZeroValue() Value {
	return t.zeroValue
}

func newArrayType(elementType Type, length int, reference bool) Type {
	return &arrayType{elementType: elementType, length: length, reference: reference}
}

type arrayType struct {
	length      int
	elementType Type
	reference   bool
}

func (a *arrayType) String() string {
	return fmt.Sprintf("array %s", a.elementType)
}

func (a *arrayType) Equals(other Type) bool {
	otherArrayType, ok := other.(*arrayType)
	if !ok {
		return false
	}

	return a.elementType.Equals(otherArrayType.elementType)
}

func (a *arrayType) ZeroValue() Value {
	arr := ArrayValue(make([]Value, a.length))
	for i := 0; i < a.length; i++ {
		arr[i] = a.elementType.ZeroValue()
	}
	return arr
}

type anyType struct {
}

func (a *anyType) Equals(_ Type) bool {
	return true
}

func (a *anyType) ZeroValue() Value {
	return nil
}

func (a *anyType) String() string {
	return "any"
}

func newMapType(keyType Type, valueType Type) Type {
	return &mapType{keyType: keyType, valueType: valueType}
}

type mapType struct {
	keyType   Type
	valueType Type
}

func (m *mapType) Equals(other Type) bool {
	otherMapType, ok := other.(*mapType)
	if !ok {
		return false
	}

	return m.keyType.Equals(otherMapType.keyType) &&
		m.valueType.Equals(otherMapType.valueType)
}

func (m *mapType) ZeroValue() Value {
	return MapValue{}
}

func (m *mapType) String() string {
	return fmt.Sprintf("map %s %s", m.keyType, m.valueType)
}

type voidType struct {
}

func (v *voidType) Equals(other Type) bool {
	return v == other
}

func (v *voidType) ZeroValue() Value {
	panic("cannot get zero value of void type")
}

func (v *voidType) String() string {
	return "void"
}

func newFunctionType(paramTypes []ParamType, returnType Type) Type {
	return &functionType{paramTypes: paramTypes, returnType: returnType}
}

type functionType struct {
	paramTypes []ParamType
	returnType Type
}

type ParamType struct {
	Name Token
	Type Type
}

func (f *functionType) Equals(other Type) bool {
	otherFunctionType, ok := other.(*functionType)
	if !ok {
		return false
	}
	if f.returnType != otherFunctionType.returnType {
		return false
	}
	if len(f.paramTypes) != len(otherFunctionType.paramTypes) {
		return false
	}
	for i, paramType := range f.paramTypes {
		if !paramType.Type.Equals(otherFunctionType.paramTypes[i].Type) {
			return false
		}
	}
	return true
}

func (f *functionType) ZeroValue() Value {
	return nil
}

func (f *functionType) String() string {
	params := make([]string, len(f.paramTypes))
	for i, paramType := range (f).paramTypes {
		params[i] = paramType.Type.String()
	}
	return fmt.Sprintf("function %s ( %s )", f.returnType, strings.Join(params, ", "))
}
