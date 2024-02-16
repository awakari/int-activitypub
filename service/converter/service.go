package converter

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/google/uuid"
)

type Service interface {
	Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error)
}

type service struct {
}

const CloudEventSpecVersion = "1.0"

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
		Source:      activity.ID.String(),
		SpecVersion: CloudEventSpecVersion,
		Type:        string(activity.Type),
		Attributes:  make(map[string]*pb.CloudEventAttributeValue),
		Data:        nil,
	}
	return
}
