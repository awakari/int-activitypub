package converter

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Service interface {
	ConvertActivityToEvent(ctx context.Context, actor vocab.Actor, activity vocab.Activity, tags util.ActivityTags) (evt *pb.CloudEvent, err error)
	ConvertEventToActivity(ctx context.Context, evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time) (a vocab.Activity, err error)
	ConvertEventToActorUpdate(ctx context.Context, evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time) (a vocab.Activity, err error)
}

type service struct {
	ceType    string
	urlBase   string
	actorType vocab.ActivityVocabularyType
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
const CeKeyName = "name"
const CeKeyObject = "object"
const CeKeyObjectUrl = "objecturl"
const CeKeyPreview = "preview"
const CeKeySrcImageUrl = "sourceimageurl"
const CeKeyStarts = "starts"
const CeKeySubject = "subject"
const CeKeySummary = "summary"
const CeKeyTime = "time"
const CeKeyTitle = "title"
const CeKeyTo = "to"
const CeKeyUpdated = "updated"

const asPublic = "https://www.w3.org/ns/activitystreams#Public"

const fmtLenMaxAttrVal = 80
const fmtLenMaxBodyTxt = 250

const ceTypePrefixFollowersOnly = "com_awakari_mastodon_"

var ErrFail = errors.New("failed to convert")

func NewService(ceType, urlBase string, actorType vocab.ActivityVocabularyType) Service {
	return service{
		ceType:    ceType,
		urlBase:   urlBase,
		actorType: actorType,
	}
}

func (svc service) ConvertActivityToEvent(ctx context.Context, actor vocab.Actor, activity vocab.Activity, tags util.ActivityTags) (evt *pb.CloudEvent, err error) {
	//
	evt = &pb.CloudEvent{
		Id:          ksuid.New().String(),
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
		case *vocab.Question:
			publicObj, err = svc.convertQuestion(objT, evt)
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
	for _, t := range tags.Object.Tag {
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

func (svc service) convertQuestion(obj *vocab.Question, evt *pb.CloudEvent) (public bool, err error) {
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

func (svc service) ConvertEventToActivity(ctx context.Context, evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time) (a vocab.Activity, err error) {

	svc.initActivity(evt, interestId, follower, t, &a)
	if evt.Type != svc.ceType && !strings.HasPrefix(evt.Type, ceTypePrefixFollowersOnly) {
		a.To = append(a.To, vocab.IRI(asPublic))
	}

	var txt string
	attrTitle, titlePresent := evt.Attributes[CeKeyTitle]
	if titlePresent {
		txt = "<b>" + attrTitle.GetCeString() + "</b><br/>"
	}
	attrSummary, summaryPresent := evt.Attributes[CeKeySummary]
	if summaryPresent {
		if txt != "" {
			txt += "<br/>"
		}
		txt += attrSummary.GetCeString()
	}
	if evt.GetTextData() != "" {
		if txt != "" {
			txt += "<br/>"
		}
		txt += evt.GetTextData()
	}
	if txt == "" {
		attrName, namePresent := evt.Attributes[CeKeyName]
		if namePresent {
			txt = attrName.GetCeString() + "<br/>"
		}
	}
	txt = truncateStringUtf8(txt, fmtLenMaxBodyTxt)

	var ceObj string
	var objType vocab.ActivityVocabularyType
	attrObj, objPresent := evt.Attributes[CeKeyObject]
	if objPresent {
		ceObj = attrObj.GetCeString()
		if ceObj == "" {
			ceObj = attrObj.GetCeUri()
		}
	}
	switch vocab.ObjectTypes.Contains(vocab.ActivityVocabularyType(ceObj)) {
	case true:
		objType = vocab.ActivityVocabularyType(ceObj)
	default:
		objType = vocab.NoteType
	}
	obj := vocab.ObjectNew(objType)
	a.Object = obj
	obj.ID = vocab.ID(svc.urlBase + "/" + evt.Id)
	switch {
	case strings.HasPrefix("http://", ceObj):
		fallthrough
	case strings.HasPrefix("https://", ceObj):
		obj.URL = vocab.IRI(ceObj)
	default:
		obj.URL = obj.ID
	}
	attrObjUrl, attrObjUrlPresent := evt.Attributes[CeKeyObjectUrl]
	if attrObjUrlPresent {
		objUrl := attrObjUrl.GetCeString()
		if objUrl == "" {
			objUrl = attrObjUrl.GetCeUri()
		}
		obj.URL = vocab.IRI(objUrl)
	}
	obj.AttributedTo = vocab.IRI(evt.Source)
	txt += fmt.Sprintf(
		"<p><a href=\"%s\">%s</a></p><p>id: %s<br/>source: %s<br/>type: %s<br/>",
		obj.URL.GetLink().String(),
		obj.URL.GetLink().String(),
		evt.Id,
		evt.Source,
		evt.Type,
	)
	var attrNames []string
	for attrName, _ := range evt.Attributes {
		attrNames = append(attrNames, attrName)
	}
	sort.Strings(attrNames)
	for _, attrName := range attrNames {
		switch attrName {
		case "awakarimatchfound": // internal
		case "awakariuserid": // do not expose
		case "awkinternal": // internal
		case CeKeyCategories:
			attrCats, _ := evt.Attributes[CeKeyCategories]
			cats := strings.Split(attrCats.GetCeString(), " ")
			var catsFormatted []string
			for _, cat := range cats {
				var tagName string
				switch strings.HasPrefix(cat, "#") {
				case true:
					tagName = cat[1:]
				default:
					tagName = cat
				}
				if len(tagName) > 0 {
					tag := vocab.LinkNew("", "")
					tag.Name = vocab.DefaultNaturalLanguageValue("#" + tagName)
					tag.Type = "Hashtag"
					tag.Href = vocab.IRI("https://mastodon.social/tags/" + tagName)
					obj.Tag = append(obj.Tag, tag)
					catFormatted := fmt.Sprintf(
						"<a rel=\"tag\" class=\"mention hashtag status-link\" href=\"https://mastodon.social/tags/%s\">%s</a>",
						tagName, tagName,
					)
					catsFormatted = append(catsFormatted, catFormatted)
				}
			}
			txt += fmt.Sprintf("%s: %s<br/>", CeKeyCategories, strings.Join(catsFormatted, " "))
		default:
			attrVal := evt.Attributes[attrName]
			switch vt := attrVal.Attr.(type) {
			case *pb.CloudEventAttributeValue_CeBoolean:
				switch vt.CeBoolean {
				case true:
					txt += fmt.Sprintf("%s: true<br/>", attrName)
				default:
					txt += fmt.Sprintf("%s: false<br/>", attrName)
				}
			case *pb.CloudEventAttributeValue_CeInteger:
				txt += fmt.Sprintf("%s: %d<br/>", attrName, vt.CeInteger)
			case *pb.CloudEventAttributeValue_CeString:
				if vt.CeString != evt.Source { // "object"/"objecturl" might the same value as the source
					txt += truncateStringUtf8(fmt.Sprintf("%s: %s", attrName, vt.CeString), fmtLenMaxAttrVal)
					txt += "<br/>"
				}
			case *pb.CloudEventAttributeValue_CeUri:
				txt += truncateStringUtf8(fmt.Sprintf("%s: %s", attrName, vt.CeUri), fmtLenMaxAttrVal)
				txt += "<br/>"
			case *pb.CloudEventAttributeValue_CeUriRef:
				txt += truncateStringUtf8(fmt.Sprintf("%s: %s", attrName, vt.CeUriRef), fmtLenMaxAttrVal)
				txt += "<br/>"
			case *pb.CloudEventAttributeValue_CeTimestamp:
				v := vt.CeTimestamp
				txt += fmt.Sprintf("%s: %s<br/>", attrName, v.AsTime().Format(time.RFC3339))
			case *pb.CloudEventAttributeValue_CeBytes:
				v := base64.StdEncoding.EncodeToString(vt.CeBytes)
				txt += truncateStringUtf8(fmt.Sprintf("%s: %s", attrName, v), fmtLenMaxAttrVal)
				txt += "<br/>"
			}
		}
	}
	txt += "</p>"
	obj.Content = vocab.DefaultNaturalLanguageValue(txt)

	if follower != nil {
		followerMention := "@" + follower.PreferredUsername.First().Value.String()
		followerUrl, _ := url.Parse(follower.URL.GetLink().String())
		if followerUrl != nil {
			followerMention += "@" + followerUrl.Hostname()
		}
		tMention := vocab.MentionNew("")
		tMention.Name = vocab.DefaultNaturalLanguageValue(followerMention)
		tMention.Href = follower.ID
		obj.Tag = append(obj.Tag, tMention)
	}

	attrTs, tsPresent := evt.Attributes[CeKeyTime]
	if tsPresent {
		obj.Published = attrTs.GetCeTimestamp().AsTime()
	}
	attrTsUpd, tsUpdPresent := evt.Attributes[CeKeyUpdated]
	if tsUpdPresent {
		obj.Updated = attrTsUpd.GetCeTimestamp().AsTime()
	}
	obj.To = a.To
	obj.CC = a.CC
	replies := vocab.CollectionNew(obj.ID + "/replies")
	obj.Replies = replies
	repliesPageFirst := vocab.CollectionPageNew(replies)
	replies.First = repliesPageFirst
	repliesPageFirst.Next = repliesPageFirst.PartOf

	attachments := vocab.ItemCollection{}
	attrIco, attrIcoPresent := evt.Attributes[CeKeyIcon]
	if attrIcoPresent {
		icoUrl := attrIco.GetCeString()
		if icoUrl == "" {
			icoUrl = attrIco.GetCeUri()
		}
		obj.Icon = vocab.LinkNew(vocab.ID(icoUrl), vocab.LinkType)
	}
	for _, k := range []string{CeKeyPreview, CeKeySrcImageUrl, CeKeyImageUrl} {
		attrImg, attrImgPresent := evt.Attributes[k]
		if attrImgPresent {
			imgUrl := attrImg.GetCeString()
			if imgUrl == "" {
				imgUrl = attrIco.GetCeUri()
			}
			obj.Image = vocab.LinkNew(vocab.ID(imgUrl), vocab.LinkType)
			attachments = append(attachments, &vocab.Object{
				Type: vocab.ImageType,
				URL:  vocab.IRI(imgUrl),
			})
		}
	}
	attrAttType, attrAttTypePresent := evt.Attributes[CeKeyAttachmentType]
	attrAttUrl, attrAttUrlPresent := evt.Attributes[CeKeyAttachmentUrl]
	if attrAttTypePresent && attrAttUrlPresent {
		objAttUrl := attrAttUrl.GetCeString()
		if objAttUrl == "" {
			objAttUrl = attrAttUrl.GetCeUri()
		}
		attachments = append(attachments, &vocab.Object{
			Type:      vocab.DocumentType,
			MediaType: vocab.MimeType(attrAttType.GetCeString()),
			URL:       vocab.IRI(objAttUrl),
		})
	}
	obj.Attachment = attachments

	attrAction, actionPresent := evt.Attributes[CeKeyAction]
	switch actionPresent {
	case true:
		a.Type = vocab.ActivityVocabularyType(attrAction.GetCeString())
	default:
		a.Type = vocab.CreateType
	}

	return
}

func (svc service) ConvertEventToActorUpdate(ctx context.Context, evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time) (a vocab.Activity, err error) {
	a = vocab.Update{
		Summary: vocab.DefaultNaturalLanguageValue(evt.GetTextData()),
		Type:    vocab.UpdateType,
	}
	svc.initActivity(evt, interestId, follower, t, &a)
	a.ID = a.ID + "-update"
	a.To = append(a.To, vocab.IRI(asPublic))
	a.Object = a.Actor
	return
}

func (svc service) initActivity(evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time, a *vocab.Activity) {
	a.ID = vocab.ID(svc.urlBase + "/" + evt.Id)
	a.Context = vocab.IRI(model.NsAs)
	a.Actor = vocab.ID(fmt.Sprintf("%s/actor/%s", svc.urlBase, interestId))
	switch t {
	case nil:
		a.Published = time.Now().UTC()
	default:
		a.Published = *t
	}
	a.To = vocab.ItemCollection{}
	if follower != nil {
		a.To = append(a.To, follower.ID)
	}
	return
}

func truncateStringUtf8(s string, lenMax int) string {
	if len(s) <= lenMax {
		return s
	}
	// Ensure we don't split a UTF-8 character in the middle.
	for i := lenMax - 3; i > 0; i-- {
		if utf8.RuneStart(s[i]) {
			return s[:i] + "..."
		}
	}
	return ""
}
