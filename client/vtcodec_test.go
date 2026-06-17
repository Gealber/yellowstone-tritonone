package client

import (
	"bytes"
	"testing"

	pb "github.com/Gealber/yellowstone-tritonone/proto"
	"google.golang.org/grpc/mem"
)

// TestVTCodecPooledRoundTrip exercises the same path the receive loop uses:
// marshal a message, Unmarshal it (via the vtproto fast path) into an object
// pulled from the pool, verify the contents, return it to the pool, then do it
// again to make sure a recycled object decodes cleanly.
func TestVTCodecPooledRoundTrip(t *testing.T) {
	codec := newVTProtoCodec()

	src := &pb.SubscribeUpdate{
		Filters: []string{"transactions_sub"},
		UpdateOneof: &pb.SubscribeUpdate_Transaction{
			Transaction: &pb.SubscribeUpdateTransaction{
				Slot: 12345,
				Transaction: &pb.SubscribeUpdateTransactionInfo{
					Signature: []byte("a-signature"),
					IsVote:    false,
					Index:     7,
				},
			},
		},
	}

	wire, err := codec.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	raw := wire.Materialize()

	decode := func() {
		buf := mem.BufferSlice{mem.SliceBuffer(append([]byte(nil), raw...))}
		msg := pb.SubscribeUpdateFromVTPool()
		defer msg.ReturnToVTPool()

		if err := codec.Unmarshal(buf, msg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		txn := msg.GetTransaction()
		if txn == nil {
			t.Fatal("expected transaction update, got nil")
		}
		if txn.Slot != 12345 {
			t.Errorf("slot = %d, want 12345", txn.Slot)
		}
		if got := txn.GetTransaction().GetSignature(); !bytes.Equal(got, []byte("a-signature")) {
			t.Errorf("signature = %q, want %q", got, "a-signature")
		}
	}

	// Run twice: the second decode reuses a recycled object from the pool.
	decode()
	decode()
}

// TestVTCodecName guards the content-subtype: it must stay "proto" so the
// server sees standard protobuf framing.
func TestVTCodecName(t *testing.T) {
	if got := newVTProtoCodec().Name(); got != "proto" {
		t.Errorf("Name() = %q, want %q", got, "proto")
	}
}
