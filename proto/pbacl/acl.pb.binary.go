// Code generated by protoc-gen-go-binary. DO NOT EDIT.
// source: pbacl/acl.proto

package pbacl

import (
	"github.com/golang/protobuf/proto"
)

// MarshalBinary implements encoding.BinaryMarshaler
func (msg *ACLLink) MarshalBinary() ([]byte, error) {
	return proto.Marshal(msg)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (msg *ACLLink) UnmarshalBinary(b []byte) error {
	return proto.Unmarshal(b, msg)
}
