package converter

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
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
const CeKeyAuthor = "author"
const CeKeyCategories = "categories"
const CeKeyCc = "cc"
const CeKeyImage = "image"
const CeKeyInReplyTo = "inreplyto"
const CeKeyLatitude = "latitude"
const CeKeyLongitude = "longitude"
const CeKeyPreview = "preview"
const CeKeySubject = "subject"
const CeKeySummary = "summary"
const CeKeyTime = "time"

const fmtAuthor = "<a href=\"%s\">%s</a>"

func NewService() Service {
	return service{}
}

func (svc service) Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
	// mastodon produces too much self delete activities, skip these
	t := activity.Type
	if t == vocab.DeleteType && activity.Actor == activity.Object {
		return
	}
	//
	evt, err = svc.convertActivity(ctx, actor, activity)
	//
	return
}

func (svc service) convertActivity(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
	//
	evt = &pb.CloudEvent{
		Id:          uuid.NewString(),
		Source:      actor.GetID().String(),
		SpecVersion: CeSpecVersion,
		Type:        CeType,
		Attributes: map[string]*pb.CloudEventAttributeValue{
			CeKeyAuthor: {
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
	//
	t := string(activity.Type)
	switch {
	case activity.Object != nil && activity.Object.IsObject():
		evt.Attributes[CeKeyAction] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: t,
			},
		}
		err = svc.convertObject(ctx, activity.Object.(*vocab.Object), evt)
	default:
		evt.Attributes[CeKeySubject] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: t,
			},
		}
		// TODO
	}
	//
	return
}

func (svc service) convertObject(ctx context.Context, obj *vocab.Object, evt *pb.CloudEvent) (err error) {
	evt.Attributes[CeKeySubject] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: string(obj.Type),
		},
	}
	if att := obj.Attachment; att != nil && att.IsObject() {
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
	}
	if aud := obj.Audience; aud != nil && len(aud) > 0 {
		var auds []string
		for _, audItem := range aud {
			if audItem.IsObject() {
				audName := audItem.(*vocab.Object).Name.String()
				auds = append(auds, audName)
			}
		}
		evt.Attributes[CeKeyAudience] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(auds, " "),
			},
		}
	}
	if cc := obj.CC; cc != nil && len(cc) > 0 {
		var ccs []string
		for _, ccItem := range cc {
			var ccStr string
			switch {
			case ccItem.IsLink():
				ccStr = ccItem.GetLink().String()
			case ccItem.IsObject():
				ccStr = ccItem.(*vocab.Object).Name.String()
			default:
				// TODO err
			}
			if ccStr != "" {
				ccs = append(ccs, ccStr)
			}
		}
		evt.Attributes[CeKeyCc] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(ccs, " "),
			},
		}
	}
	if img := obj.Image; img != nil {
		evt.Attributes[CeKeyImage] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: img.GetLink().String(),
			},
		}
	}
	if inReplyTo := obj.InReplyTo; inReplyTo != nil {
		switch {
		case inReplyTo.IsLink():
			evt.Attributes[CeKeyImage] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeUri{
					CeUri: inReplyTo.GetLink().String(),
				},
			}
		case inReplyTo.IsObject():
			evt.Attributes[CeKeyImage] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: inReplyTo.(*vocab.Object).Name.String(),
				},
			}
		default:
			// TODO err
		}
	}
	if loc := obj.Location; loc != nil {
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
			// TODO err
		}
	}
	if preview := obj.Preview; preview != nil {
		evt.Attributes[CeKeyPreview] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: preview.GetLink().String(),
			},
		}
	}
	if summ := obj.Summary; summ != nil && len(summ) > 0 {
		evt.Attributes[CeKeySummary] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: summ.String(),
			},
		}
	}
	if tags := obj.Tag; tags != nil && len(tags) > 0 {
		var cats []string
		for _, tag := range tags {
			switch tagT := tag.(type) {
			case *vocab.Object:
				cats = append(cats, tagT.Name.String())
			default:
				// TODO: err
			}
		}
		evt.Attributes[CeKeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(cats, " "),
			},
		}
	}
	evt.Data = &pb.CloudEvent_TextData{
		TextData: obj.Content.String(),
	}
	return
}
