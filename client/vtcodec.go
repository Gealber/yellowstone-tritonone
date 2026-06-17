package client

import (
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/mem"
)

// vtMarshaler / vtUnmarshaler are implemented by messages generated with
// protoc-gen-go-vtproto (features=marshal+unmarshal).
type vtMarshaler interface {
	MarshalVT() ([]byte, error)
}

type vtUnmarshaler interface {
	UnmarshalVT([]byte) error
}

// vtCodecName matches the standard "proto" content-subtype: vtprotobuf emits
// the exact same protobuf wire format, so the server needs no special handling.
const vtCodecName = "proto"

// vtprotoCodec is a gRPC CodecV2 that uses vtprotobuf's generated
// MarshalVT/UnmarshalVT fast paths when the message implements them, and falls
// back to the standard proto codec for everything else (e.g. health checks).
type vtprotoCodec struct {
	fallback encoding.CodecV2
}

func newVTProtoCodec() vtprotoCodec {
	return vtprotoCodec{fallback: encoding.GetCodecV2(vtCodecName)}
}

func (c vtprotoCodec) Marshal(v any) (mem.BufferSlice, error) {
	if m, ok := v.(vtMarshaler); ok {
		data, err := m.MarshalVT()
		if err != nil {
			return nil, err
		}
		return mem.BufferSlice{mem.SliceBuffer(data)}, nil
	}
	return c.fallback.Marshal(v)
}

func (c vtprotoCodec) Unmarshal(data mem.BufferSlice, v any) error {
	if m, ok := v.(vtUnmarshaler); ok {
		// data is freed once this returns; UnmarshalVT copies what it needs.
		return m.UnmarshalVT(data.Materialize())
	}
	return c.fallback.Unmarshal(data, v)
}

func (vtprotoCodec) Name() string { return vtCodecName }
