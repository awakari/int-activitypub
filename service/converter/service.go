package converter

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strings"
)

type Service interface {
	Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity, tags util.ActivityTags) (evt *pb.CloudEvent, err error)
}

type service struct {
	ceType string
}

const CeSpecVersion = "1.0"
const CeKeyAction = "action"
const CeKeyAttachmentUrl = "attachmenturl"
const CeKeyAttachmentType = "attachmenttype"
const CeKeyAudience = "audience"
const CeKeyCategories = "categories"
const CeKeyCc = "cc"
const CeKeyDuration = "duration"
const CeKeyEnds = "ends"
const CeKeyIcon = "icon"
const CeKeyImageUrl = "imageurl"
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
const CeKeyTo = "to"
const CeKeyUpdated = "updated"

const asPublic = "https://www.w3.org/ns/activitystreams#Public"

var ErrFail = errors.New("failed to convert")

func NewService(ceType string) Service {
	return service{
		ceType: ceType,
	}
}

func (svc service) Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity, tags util.ActivityTags) (evt *pb.CloudEvent, err error) {
	//
	evt = &pb.CloudEvent{
		Id:          uuid.NewString(),
		Source:      actor.ID.String(),
		SpecVersion: CeSpecVersion,
		Type:        svc.ceType,
		Attributes: map[string]*pb.CloudEventAttributeValue{
			CeKeyTime: {
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(activity.Published),
				},
			},
		},
	}
	if actor.Name.Count() > 0 {
		evt.Attributes[CeKeySubject] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: actor.Name.String(),
			},
		}
	}
	//
	var public bool
	public, err = svc.convertActivity(activity, evt, tags)
	var publicObj bool
	t := string(activity.Type)
	if activity.Object != nil {
		evt.Attributes[CeKeyAction] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: t,
			},
		}
		obj := activity.Object
		switch objT := obj.(type) {
		case *vocab.Object:
			publicObj, err = svc.convertObject(objT, evt)
		default:
			switch obj.IsLink() {
			case true:
				evt.Attributes[CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeUri{
						CeUri: obj.GetLink().String(),
					},
				}
			default:
				err = fmt.Errorf("%w activity object, unexpected type: %s", ErrFail, reflect.TypeOf(obj))
			}
		}
	}
	// honor the privacy: discard any publication that is not explicitly public
	if !public && !publicObj {
		evt = nil
	}
	//
	return
}

func (svc service) convertActivity(a vocab.Activity, evt *pb.CloudEvent, tags util.ActivityTags) (public bool, err error) {
	evt.Attributes[CeKeyObject] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: string(a.Type),
		},
	}
	evt.Attributes[CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeUri{
			CeUri: a.ID.String(),
		},
	}
	if att := a.Attachment; att != nil {
		err = convertAttachment(att, evt)
	}
	if aud := a.Audience; aud != nil && len(aud) > 0 {
		var publicAud bool
		var errAud error
		publicAud, errAud = convertAsCollectionDetectAsPublic(aud, evt, CeKeyAudience)
		if publicAud {
			public = true
		}
		err = errors.Join(err, errAud)
	}
	if cc := a.CC; cc != nil && len(cc) > 0 {
		var publicCc bool
		var errCc error
		publicCc, errCc = convertAsCollectionDetectAsPublic(cc, evt, CeKeyCc)
		if publicCc {
			public = true
		}
		err = errors.Join(err, errCc)
	}
	if a.Content != nil {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: a.Content.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", a.Content.String(), txt),
			}
		}
	}
	if a.Duration > 0 {
		evt.Attributes[CeKeyDuration] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: int32(a.Duration.Seconds()),
			},
		}
	}
	if !a.EndTime.IsZero() {
		evt.Attributes[CeKeyEnds] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(a.EndTime),
			},
		}
	}
	if ico := a.Icon; ico != nil {
		err = errors.Join(err, convertAsLink(ico, evt, CeKeyIcon))
	}
	if img := a.Image; img != nil {
		err = errors.Join(err, convertAsLink(img, evt, CeKeyImageUrl))
	}
	if inReplyTo := a.InReplyTo; inReplyTo != nil {
		err = errors.Join(err, convertInReplyTo(inReplyTo, evt))
	}
	if loc := a.Location; loc != nil {
		err = errors.Join(err, convertLocation(loc, evt))
	}
	if preview := a.Preview; preview != nil {
		err = errors.Join(err, convertAsLink(preview, evt, CeKeyPreview))
	}
	if !a.StartTime.IsZero() {
		evt.Attributes[CeKeyStarts] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(a.StartTime),
			},
		}
	}
	if summ := a.Summary; summ != nil && len(summ) > 0 {
		txt := evt.GetTextData()
		switch txt {
		case "":
			evt.Data = &pb.CloudEvent_TextData{
				TextData: summ.String(),
			}
		default:
			evt.Data = &pb.CloudEvent_TextData{
				TextData: fmt.Sprintf("%s\n\n%s", summ.String(), txt),
			}
		}
		err = errors.Join(err, convertAsText(summ, evt, CeKeySummary))
	}
	var tagNames []string
	for _, t := range tags.Tag {
		tagNames = append(tagNames, t.Name)
	}
	if len(tagNames) > 0 {
		evt.Attributes[CeKeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(tagNames, " "),
			},
		}
	}
	if to := a.To; to != nil && len(to) > 0 {
		var publicTo bool
		var errTo error
		publicTo, errTo = convertAsCollectionDetectAsPublic(to, evt, CeKeyTo)
		if publicTo {
			public = true
		}
		err = errors.Join(err, errTo)
	}
	if !a.Updated.IsZero() {
		evt.Attributes[CeKeyUpdated] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(a.Updated),
			},
		}
	}
	return
}

func (svc service) convertObject(obj *vocab.Object, evt *pb.CloudEvent) (public bool, err error) {
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
		var publicAud bool
		var errAud error
		publicAud, errAud = convertAsCollectionDetectAsPublic(aud, evt, CeKeyAudience)
		if publicAud {
			public = true
		}
		err = errors.Join(err, errAud)
	}
	if cc := obj.CC; cc != nil && len(cc) > 0 {
		var publicCc bool
		var errCc error
		publicCc, errCc = convertAsCollectionDetectAsPublic(cc, evt, CeKeyCc)
		if publicCc {
			public = true
		}
		err = errors.Join(err, errCc)
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
				TextData: fmt.Sprintf("%s\n\n%s", obj.Content.String(), txt),
			}
		}
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
		err = errors.Join(err, convertAsLink(img, evt, CeKeyImageUrl))
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
				TextData: fmt.Sprintf("%s\n\n%s", summ.String(), txt),
			}
		}
	}
	if tags := obj.Tag; tags != nil && len(tags) > 0 {
		err = errors.Join(err, convertAsCollection(tags, evt, CeKeyCategories))
	}
	if to := obj.To; to != nil && len(to) > 0 {
		var publicTo bool
		var errTo error
		publicTo, errTo = convertAsCollectionDetectAsPublic(to, evt, CeKeyTo)
		if publicTo {
			public = true
		}
		err = errors.Join(err, errTo)
	}
	if !obj.Updated.IsZero() {
		evt.Attributes[CeKeyUpdated] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(obj.Updated),
			},
		}
	}
	//
	return
}

func convertAsCollectionDetectAsPublic(items vocab.ItemCollection, evt *pb.CloudEvent, key string) (public bool, err error) {
	var result []string
	for _, item := range items {
		var itemStr string
		switch {
		case item.IsLink():
			itemStr = item.GetLink().String()
		case item.IsObject():
			itemStr = item.(*vocab.Object).Name.String()
		default:
			err = fmt.Errorf("%w item in the collection, unexpected type: %s", ErrFail, reflect.TypeOf(item))
		}
		if itemStr != "" {
			result = append(result, itemStr)
		}
		if itemStr == asPublic {
			public = true
		}
	}
	evt.Attributes[key] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: strings.Join(result, " "),
		},
	}
	return
}

func convertAsCollection(items vocab.ItemCollection, evt *pb.CloudEvent, key string) (err error) {
	_, err = convertAsCollectionDetectAsPublic(items, evt, key)
	return
}

func convertAttachment(att vocab.Item, evt *pb.CloudEvent) (err error) {
	switch {
	case att.IsLink():
		evt.Attributes[CeKeyAttachmentUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: att.GetLink().String(),
			},
		}
	case att.IsObject():
		attObj := att.(*vocab.Object)
		evt.Attributes[CeKeyAttachmentUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: attObj.URL.GetLink().String(),
			},
		}
		evt.Attributes[CeKeyAttachmentType] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: string(attObj.MediaType),
			},
		}
	case att.IsCollection():
		attColl := att.(vocab.ItemCollection)
		if attColl.Count() > 0 {
			err = convertAttachment(attColl.First(), evt)
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
	if len(item) > 0 {
		evt.Attributes[key] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.String(),
			},
		}
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
