package converter

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strings"
)

type Service interface {
	Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error)
}

type service struct {
}

const CeType = "com.awakari.activitypub.v1"
const CeSpecVersion = "1.0"
const CeKeyAction = "action"
const CeKeyAttachment = "attachment"
const CeKeyAttachmentType = "attachmenttype"
const CeKeyAudience = "audience"
const CeKeyCategories = "categories"
const CeKeyCc = "cc"
const CeKeyDuration = "duration"
const CeKeyEnds = "ends"
const CeKeyIcon = "icon"
const CeKeyImage = "image"
const CeKeyInReplyTo = "inreplyto"
const CeKeyLatitude = "latitude"
const CeKeyLongitude = "longitude"
const CeKeyObject = "object"
const CeKeyObjectUrl = "objecturl"
const CeKeyPreview = "preview"
const CeKeyStarts = "starts"
const CeKeySubject = "subject"
const CeKeySummary = "summary"
const CeKeyTime = "time"
const CeKeyUpdated = "updated"

var ErrFail = errors.New("failed to convert")

func NewService() Service {
	return service{}
}

func (svc service) Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
	t := activity.Type
	if t == "" || t == vocab.DeleteType && activity.Actor == activity.Object {
		return // skip self delete activities w/o an error
	}
	evt, err = svc.convertActivity(actor, activity)
	return
}

func (svc service) convertActivity(actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
	//
	evt = &pb.CloudEvent{
		Id:          uuid.NewString(),
		Source:      actor.ID.String(),
		SpecVersion: CeSpecVersion,
		Type:        CeType,
		Attributes: map[string]*pb.CloudEventAttributeValue{
			CeKeySubject: {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: actor.Name.String(),
				},
			},
			CeKeyTime: {
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(activity.Published),
				},
			},
		},
	}
	if activity.Content != nil {
		evt.Data = &pb.CloudEvent_TextData{
			TextData: activity.Content.String(),
		}
	}
	if activity.Summary != nil {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: activity.Summary.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", activity.Summary, txt),
			}
		}
	}
	//
	t := string(activity.Type)
	switch {
	case activity.Object != nil:
		evt.Attributes[CeKeyAction] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: t,
			},
		}
		obj := activity.Object
		switch {
		case obj.IsLink():
			evt.Attributes[CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeUri{
					CeUri: obj.GetLink().String(),
				},
			}
		case obj.IsObject():
			err = svc.convertObject(obj.(*vocab.Object), evt)
		default:
			err = fmt.Errorf("%w activity object, unexpected type: %s", ErrFail, reflect.TypeOf(obj))
		}
	default:
		err = svc.convertActivityAsObject(activity, evt)
	}
	//
	return
}

func (svc service) convertObject(obj *vocab.Object, evt *pb.CloudEvent) (err error) {
	evt.Attributes[CeKeyObject] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: string(obj.Type),
		},
	}
	evt.Attributes[CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeUri{
			CeUri: obj.ID.String(),
		},
	}
	if att := obj.Attachment; att != nil {
		err = convertAttachment(att, evt)
	}
	if aud := obj.Audience; aud != nil && len(aud) > 0 {
		err = errors.Join(err, convertAsCollection(aud, evt, CeKeyAudience))
	}
	if cc := obj.CC; cc != nil && len(cc) > 0 {
		err = errors.Join(err, convertAsCollection(cc, evt, CeKeyCc))
	}
	if obj.Duration > 0 {
		evt.Attributes[CeKeyDuration] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: int32(obj.Duration.Seconds()),
			},
		}
	}
	if !obj.EndTime.IsZero() {
		evt.Attributes[CeKeyEnds] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.EndTime),
			},
		}
	}
	if ico := obj.Icon; ico != nil {
		err = errors.Join(err, convertAsLink(ico, evt, CeKeyIcon))
	}
	if img := obj.Image; img != nil {
		err = errors.Join(err, convertAsLink(img, evt, CeKeyImage))
	}
	if inReplyTo := obj.InReplyTo; inReplyTo != nil {
		err = errors.Join(err, convertInReplyTo(inReplyTo, evt))
	}
	if loc := obj.Location; loc != nil {
		err = errors.Join(err, convertLocation(loc, evt))
	}
	if preview := obj.Preview; preview != nil {
		err = errors.Join(err, convertAsLink(preview, evt, CeKeyPreview))
	}
	if !obj.StartTime.IsZero() {
		evt.Attributes[CeKeyStarts] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.StartTime),
			},
		}
	}
	if summ := obj.Summary; summ != nil && len(summ) > 0 {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: summ.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", summ, txt),
			}
		}
	}
	if tags := obj.Tag; tags != nil && len(tags) > 0 {
		err = errors.Join(err, convertAsCollection(tags, evt, CeKeyCategories))
	}
	if !obj.Updated.IsZero() {
		evt.Attributes[CeKeyUpdated] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.Updated),
			},
		}
	}
	if obj.Content != nil {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: obj.Content.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", obj.Content, txt),
			}
		}
	}
	return
}

func (svc service) convertActivityAsObject(obj vocab.Activity, evt *pb.CloudEvent) (err error) {
	evt.Attributes[CeKeyObject] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: string(obj.Type),
		},
	}
	evt.Attributes[CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeUri{
			CeUri: obj.ID.String(),
		},
	}
	if att := obj.Attachment; att != nil {
		err = convertAttachment(att, evt)
	}
	if aud := obj.Audience; aud != nil && len(aud) > 0 {
		err = errors.Join(err, convertAsCollection(aud, evt, CeKeyAudience))
	}
	if cc := obj.CC; cc != nil && len(cc) > 0 {
		err = errors.Join(err, convertAsCollection(cc, evt, CeKeyCc))
	}
	if obj.Duration > 0 {
		evt.Attributes[CeKeyDuration] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: int32(obj.Duration.Seconds()),
			},
		}
	}
	if !obj.EndTime.IsZero() {
		evt.Attributes[CeKeyEnds] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.EndTime),
			},
		}
	}
	if ico := obj.Icon; ico != nil {
		err = errors.Join(err, convertAsLink(ico, evt, CeKeyIcon))
	}
	if img := obj.Image; img != nil {
		err = errors.Join(err, convertAsLink(img, evt, CeKeyImage))
	}
	if inReplyTo := obj.InReplyTo; inReplyTo != nil {
		err = errors.Join(err, convertInReplyTo(inReplyTo, evt))
	}
	if loc := obj.Location; loc != nil {
		err = errors.Join(err, convertLocation(loc, evt))
	}
	if preview := obj.Preview; preview != nil {
		err = errors.Join(err, convertAsLink(preview, evt, CeKeyPreview))
	}
	if !obj.StartTime.IsZero() {
		evt.Attributes[CeKeyStarts] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.StartTime),
			},
		}
	}
	if summ := obj.Summary; summ != nil && len(summ) > 0 {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: summ.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", summ, txt),
			}
		}
		err = errors.Join(err, convertAsText(summ, evt, CeKeySummary))
	}
	if tags := obj.Tag; tags != nil && len(tags) > 0 {
		err = errors.Join(err, convertAsCollection(tags, evt, CeKeyCategories))
	}
	if !obj.Updated.IsZero() {
		evt.Attributes[CeKeyUpdated] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.Updated),
			},
		}
	}
	if obj.Content != nil {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: obj.Content.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", obj.Content, txt),
			}
		}
	}
	return
}

func convertAsCollection(cc vocab.ItemCollection, evt *pb.CloudEvent, key string) (err error) {
	var ccs []string
	for _, ccItem := range cc {
		var ccStr string
		switch {
		case ccItem.IsLink():
			ccStr = ccItem.GetLink().String()
		case ccItem.IsObject():
			ccStr = ccItem.(*vocab.Object).Name.String()
		default:
			err = fmt.Errorf("%w item in the collection, unexpected type: %s", ErrFail, reflect.TypeOf(ccItem))
		}
		if ccStr != "" {
			ccs = append(ccs, ccStr)
		}
	}
	evt.Attributes[key] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: strings.Join(ccs, " "),
		},
	}
	return
}

func convertAttachment(att vocab.Item, evt *pb.CloudEvent) (err error) {
	switch {
	case att.IsLink():
		evt.Attributes[CeKeyAttachment] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: att.GetLink().String(),
			},
		}
	case att.IsObject():
		attObj := att.(*vocab.Object)
		evt.Attributes[CeKeyAttachment] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: attObj.URL.GetLink().String(),
			},
		}
		evt.Attributes[CeKeyAttachmentType] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: string(attObj.MediaType),
			},
		}
	default:
		err = fmt.Errorf("%w attachment, unexpected type: %s", ErrFail, reflect.TypeOf(att))
	}
	return
}

func convertAsLink(item vocab.Item, evt *pb.CloudEvent, key string) (err error) {
	evt.Attributes[key] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeUri{
			CeUri: item.GetLink().String(),
		},
	}
	return
}

func convertAsText(item vocab.NaturalLanguageValues, evt *pb.CloudEvent, key string) (err error) {
	evt.Attributes[key] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeUri{
			CeUri: item.String(),
		},
	}
	return
}

func convertInReplyTo(inReplyTo vocab.Item, evt *pb.CloudEvent) (err error) {
	switch {
	case inReplyTo.IsLink():
		evt.Attributes[CeKeyInReplyTo] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: inReplyTo.GetLink().String(),
			},
		}
	case inReplyTo.IsObject():
		evt.Attributes[CeKeyInReplyTo] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: inReplyTo.(*vocab.Object).Name.String(),
			},
		}
	default:
		err = fmt.Errorf("%w \"inReplyTo\", unexpected type: %s", ErrFail, reflect.TypeOf(inReplyTo))
	}
	return
}

func convertLocation(loc vocab.Item, evt *pb.CloudEvent) (err error) {
	switch locT := loc.(type) {
	case *vocab.Place:
		evt.Attributes[CeKeyLatitude] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: fmt.Sprintf("%f", locT.Latitude),
			},
		}
		evt.Attributes[CeKeyLongitude] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: fmt.Sprintf("%f", locT.Longitude),
			},
		}
	default:
		err = fmt.Errorf("%w location, unexpected type: %s", ErrFail, reflect.TypeOf(loc))
	}
	return
}
