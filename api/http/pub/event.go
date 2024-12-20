package pub

import (
	"encoding/base64"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"time"
)

type event struct {
	Id          string               `json:"id"`
	SpecVersion string               `json:"specVersion,omitempty"`
	Source      string               `json:"source"`
	Type        string               `json:"type"`
	Attributes  map[string]attribute `json:"attributes"`
	TextData    string               `json:"textData,omitempty"`
}

type attribute struct {
	CeBoolean   *bool      `json:"ceBoolean,omitempty"`
	CeBytes     *string    `json:"ceBytes,omitempty"`
	CeInteger   *int32     `json:"ceInteger,omitempty"`
	CeString    *string    `json:"ceString,omitempty"`
	CeTimestamp *time.Time `json:"ceTimestamp,omitempty"`
	CeUri       *string    `json:"ceUri,omitempty"`
	CeUriRef    *string    `json:"ceUriRef,omitempty"`
}

func MarshalEvent(src *pb.CloudEvent) (data []byte, err error) {

	evt := event{
		Id:          src.Id,
		SpecVersion: src.SpecVersion,
		Source:      src.Source,
		Type:        src.Type,
		Attributes:  make(map[string]attribute),
		TextData:    src.GetTextData(),
	}

	for k, v := range src.GetAttributes() {
		switch vt := v.Attr.(type) {
		case *pb.CloudEventAttributeValue_CeBoolean:
			evt.Attributes[k] = attribute{
				CeBoolean: &vt.CeBoolean,
			}
		case *pb.CloudEventAttributeValue_CeBytes:
			b64s := base64.StdEncoding.EncodeToString(vt.CeBytes)
			evt.Attributes[k] = attribute{
				CeBytes: &b64s,
			}
		case *pb.CloudEventAttributeValue_CeInteger:
			evt.Attributes[k] = attribute{
				CeInteger: &vt.CeInteger,
			}
		case *pb.CloudEventAttributeValue_CeString:
			evt.Attributes[k] = attribute{
				CeString: &vt.CeString,
			}
		case *pb.CloudEventAttributeValue_CeTimestamp:
			ts := vt.CeTimestamp.AsTime().UTC()
			evt.Attributes[k] = attribute{
				CeTimestamp: &ts,
			}
		case *pb.CloudEventAttributeValue_CeUri:
			evt.Attributes[k] = attribute{
				CeUri: &vt.CeUri,
			}
		case *pb.CloudEventAttributeValue_CeUriRef:
			evt.Attributes[k] = attribute{
				CeUriRef: &vt.CeUriRef,
			}
		default:
			err = fmt.Errorf("failed to marshal event %s, unknown attribute type: %T", src.Id, vt)
		}
	}

	if err == nil {
		data, err = sonic.Marshal(evt)
	}

	return
}
