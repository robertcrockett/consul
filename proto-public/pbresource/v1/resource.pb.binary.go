// Code generated by protoc-gen-go-binary. DO NOT EDIT.
// source: pbresource/v1/resource.proto

package resourcev1

import (
	"google.golang.org/protobuf/proto"
)

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Type) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Type) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Tenancy) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Tenancy) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *ID) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *ID) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Resource) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Resource) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Status) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Status) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Condition) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Condition) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *Reference) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *Reference) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}
