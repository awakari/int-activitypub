package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type recSrc struct {
	ActorId  string `bson:"actorId"`
	GroupId  string `bson:"groupId"`
	UserId   string `bson:"userId"`
	Type     string `bson:"type"`
	Name     string `bson:"name"`
	Summary  string `bson:"summary"`
	Accepted bool   `bson:"accepted"`
}

const attrActorId = "actorId"
const attrGroupId = "groupId"
const attrUserId = "userId"
const attrType = "type"
const attrName = "name"
const attrSummary = "summary"
const attrAccepted = "accepted"

type storageMongo struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

var optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
var optsRead = options.
	FindOne().
	SetShowRecordID(false).
	SetProjection(projRead)
var projRead = bson.D{
	{
		Key:   attrActorId,
		Value: 1,
	},
	{
		Key:   attrGroupId,
		Value: 1,
	},
	{
		Key:   attrUserId,
		Value: 1,
	},
	{
		Key:   attrType,
		Value: 1,
	},
	{
		Key:   attrName,
		Value: 1,
	},
	{
		Key:   attrSummary,
		Value: 1,
	},
	{
		Key:   attrAccepted,
		Value: 1,
	},
}

var sortGetBatchAsc = bson.D{
	{
		Key:   attrActorId,
		Value: 1,
	},
}
var sortGetBatchDesc = bson.D{
	{
		Key:   attrActorId,
		Value: -1,
	},
}
var indices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{
				Key:   attrActorId,
				Value: 1,
			},
		},
		Options: options.
			Index().
			SetUnique(true),
	},
}

func NewStorage(ctx context.Context, cfgDb config.DbConfig) (s Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
	if cfgDb.Tls.Enabled {
		clientOpts = clientOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cfgDb.Tls.Insecure})
	}
	if len(cfgDb.UserName) > 0 {
		auth := options.Credential{
			Username:    cfgDb.UserName,
			Password:    cfgDb.Password,
			PasswordSet: len(cfgDb.Password) > 0,
		}
		clientOpts = clientOpts.SetAuth(auth)
	}
	conn, err := mongo.Connect(ctx, clientOpts)
	var sm storageMongo
	if err == nil {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Following.Name)
		sm.conn = conn
		sm.db = db
		sm.coll = coll
		_, err = sm.ensureIndices(ctx)
	}
	if err == nil {
		s = sm
	}
	return
}

func (sm storageMongo) ensureIndices(ctx context.Context) ([]string, error) {
	return sm.coll.Indexes().CreateMany(ctx, indices)
}

func (sm storageMongo) Close() error {
	return sm.conn.Disconnect(context.TODO())
}

func (sm storageMongo) Create(ctx context.Context, src model.Source) (err error) {
	rec := recSrc{
		ActorId: src.ActorId,
		GroupId: src.GroupId,
		UserId:  src.UserId,
		Type:    src.Type,
		Name:    src.Name,
		Summary: src.Summary,
	}
	_, err = sm.coll.InsertOne(ctx, rec)
	err = decodeError(err, src.ActorId)
	return
}

func (sm storageMongo) Read(ctx context.Context, srcId string) (a model.Source, err error) {
	q := bson.M{
		attrActorId: srcId,
	}
	var result *mongo.SingleResult
	result = sm.coll.FindOne(ctx, q, optsRead)
	err = result.Err()
	var rec recSrc
	if err == nil {
		err = result.Decode(&rec)
	}
	if err == nil {
		a.ActorId = rec.ActorId
		a.GroupId = rec.GroupId
		a.UserId = rec.UserId
		a.Type = rec.Type
		a.Name = rec.Name
		a.Summary = rec.Summary
		a.Accepted = rec.Accepted
	}
	err = decodeError(err, srcId)
	return
}

func (sm storageMongo) Update(ctx context.Context, src model.Source) (err error) {
	q := bson.M{
		attrActorId: src.ActorId,
	}
	var result *mongo.UpdateResult
	result, err = sm.coll.UpdateOne(ctx, q, optsRead)
	switch err {
	case nil:
		// TODO check
	default:
		err = decodeError(err, src.ActorId)
	}
	return
}

func (sm storageMongo) Delete(ctx context.Context, srcId, groupId, userId string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (sm storageMongo) List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	//TODO implement me
	panic("implement me")
}

func decodeError(src error, recId string) (dst error) {
	switch {
	case src == nil:
	case errors.Is(src, mongo.ErrNoDocuments):
		dst = fmt.Errorf("%w: %s", ErrNotFound, recId)
	case mongo.IsDuplicateKeyError(src):
		dst = fmt.Errorf("%w: %s", ErrConflict, recId)
	default:
		dst = fmt.Errorf("%w: %s", ErrInternal, src)
	}
	return
}
