package converter

import (
    "context"
    "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
    vocab "github.com/go-ap/activitypub"
    "github.com/google/uuid"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
    Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error)
}

type service struct {
}

const CeType = "com.awakari.activitypub.v1"
const CeSpecVersion = "1.0"
const CeKeyAction = "action"
const CeKeyAuthor = "author"
const CeKeySubject = "subject"
const CeKeyTime = "time"

const fmtAuthor = "<a href=\"%s\">%s</a>"

func NewService() Service {
    return service{}
}

func (svc service) Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
    // mastodon produces too much self delete activities, skip these
    if activity.Type == vocab.DeleteType && activity.Actor == activity.Object {
        return
    }
    //
    evt = &pb.CloudEvent{
        Id:          uuid.NewString(),
        Source:      activity.Actor.GetLink().String(),
        SpecVersion: CeSpecVersion,
        Type:        CeType,
        Attributes: map[string]*pb.CloudEventAttributeValue{
            CeKeyAction: {
                Attr: &pb.CloudEventAttributeValue_CeString{
                    CeString: string(activity.Type),
                },
            },
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
    err = svc.convertObject(ctx, activity.Object.(*vocab.Object), evt)
    //
    return
}

func (svc service) convertObject(ctx context.Context, o *vocab.Object, evt *pb.CloudEvent) (err error) {
    evt.Attributes[CeKeySubject] = &pb.CloudEventAttributeValue{
        Attr: &pb.CloudEventAttributeValue_CeString{
            CeString: string(o.Type),
        },
    }
    if att := o.Attachment; att != nil {
        evt.Attributes[CeKeyAttachment] = &pb.CloudEventAttributeValue{
            Attr: &pb.CloudEventAttributeValue_CeUri{
                CeUri: att.GetLink().String(),
            },
        }
    }
    if summ := o.Summary; summ != nil && len(summ) > 0 {
        evt.Attributes[CeKeySummary] = &pb.CloudEventAttributeValue{
            Attr: &pb.CloudEventAttributeValue_CeString{
                CeString: summ.String(),
            },
        }
    }
    if img := o.Image; img != nil {
        evt.Attributes[CeKeyImage] = &pb.CloudEventAttributeValue{
            Attr: &pb.CloudEventAttributeValue_CeUri{
                CeUri: img.GetLink().String(),
            },
        }
    }
    if loc := o.Location; loc != nil {
        evt.Attributes[CeKeyLatitude] = &pb.CloudEventAttributeValue{
            Attr: &pb.CloudEventAttributeValue_CeString{
               CeString: loc.,
            },
        }
    }
    evt.Data = &pb.CloudEvent_TextData{
        TextData: o.Content.String(),
    }
    return
}
