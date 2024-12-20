package pub

import (
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestMarshalEvent(t *testing.T) {
	in := &pb.CloudEvent{
		Id:          "id1",
		Source:      "src1",
		SpecVersion: "1.0",
		Type:        "type1",
		Attributes: map[string]*pb.CloudEventAttributeValue{
			"boolean1": {
				Attr: &pb.CloudEventAttributeValue_CeBoolean{
					CeBoolean: true,
				},
			},
			"bytes1": {
				Attr: &pb.CloudEventAttributeValue_CeBytes{
					CeBytes: []byte("some bytes"),
				},
			},
			"integer1": {
				Attr: &pb.CloudEventAttributeValue_CeInteger{
					CeInteger: -42,
				},
			},
			"string1": {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: "string1",
				},
			},
			"timestamp1": {
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(time.Date(2024, 12, 20, 10, 40, 15, 0, time.UTC)),
				},
			},
			"uri1": {
				Attr: &pb.CloudEventAttributeValue_CeUri{
					CeUri: "uri1",
				},
			},
		},
		Data: &pb.CloudEvent_TextData{
			TextData: "text1",
		},
	}
	out, err := MarshalEvent(in)
	require.NoError(t, err)
	fmt.Println(string(out))
}
