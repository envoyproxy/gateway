package utils

import (
	"bytes"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func MarshalResourcesToJSON(resources []types.Resource) ([]byte, error) {
	msgs := make([]proto.Message, 0)
	for _, resource := range resources {
		msgs = append(msgs, resource.(proto.Message))
	}
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	for idx, msg := range msgs {
		if idx != 0 {
			buffer.WriteByte(',')
		}
		b, err := protojson.Marshal(msg)
		if err != nil {
			return nil, err
		}
		buffer.Write(b)
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}
