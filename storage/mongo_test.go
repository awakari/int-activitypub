package storage

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("feeds-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Table.Following.Shard = false
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	//
	clear(ctx, t, s.(storageMongo))
}

func clear(ctx context.Context, t *testing.T, s storageMongo) {
	require.Nil(t, s.coll.Drop(ctx))
	require.Nil(t, s.Close())
}

func TestStorageMongo_Create(t *testing.T) {
	//
	collName := fmt.Sprintf("following-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(storageMongo))
	//
	err = s.Create(ctx, model.Source{
		ActorId: "actor0",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name0",
		Summary: "summary0",
		Created: time.Date(2024, 4, 11, 16, 39, 35, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		src model.Source
		err error
	}{
		"ok": {
			src: model.Source{
				ActorId: "actor1",
			},
		},
		"conflict": {
			src: model.Source{
				ActorId: "actor0",
			},
			err: ErrConflict,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = s.Create(ctx, c.src)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_Read(t *testing.T) {
	//
	collName := fmt.Sprintf("following-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(storageMongo))
	//
	err = s.Create(ctx, model.Source{
		ActorId: "actor0",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name0",
		Summary: "summary0",
		Created: time.Date(2024, 4, 11, 16, 39, 35, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		src model.Source
		err error
	}{
		"ok": {
			id: "actor0",
			src: model.Source{
				ActorId: "actor0",
				GroupId: "group0",
				UserId:  "user0",
				Type:    "type0",
				Name:    "name0",
				Summary: "summary0",
				Created: time.Date(2024, 4, 11, 16, 39, 35, 0, time.UTC),
			},
		},
		"missing": {
			id:  "actor1",
			err: ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			src, err := s.Read(ctx, c.id)
			assert.Equal(t, c.src, src)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_Update(t *testing.T) {
	//
	collName := fmt.Sprintf("following-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(storageMongo))
	//
	err = s.Create(ctx, model.Source{
		ActorId: "actor0",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name0",
		Summary: "summary0",
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		src model.Source
		err error
	}{
		"ok": {
			src: model.Source{
				ActorId:  "actor0",
				GroupId:  "group0",
				UserId:   "user0",
				Type:     "type0",
				Name:     "name0",
				Summary:  "summary0",
				Accepted: true,
			},
		},
		"missing": {
			src: model.Source{
				ActorId:  "actor1",
				GroupId:  "group0",
				UserId:   "user0",
				Type:     "type0",
				Name:     "name0",
				Summary:  "summary0",
				Accepted: true,
			},
			err: ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = s.Update(ctx, c.src)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_Delete(t *testing.T) {
	//
	collName := fmt.Sprintf("following-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(storageMongo))
	//
	err = s.Create(ctx, model.Source{
		ActorId: "actor0",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name0",
		Summary: "summary0",
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		err     error
	}{
		"ok": {
			id:      "actor0",
			groupId: "group0",
			userId:  "user0",
		},
		"missing1": {
			id:      "actor1",
			groupId: "group0",
			userId:  "user0",
			err:     ErrNotFound,
		},
		"missing2": {
			id:      "actor0",
			groupId: "group0",
			userId:  "user1",
			err:     ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = s.Delete(ctx, c.id, c.groupId, c.userId)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_ListUrls(t *testing.T) {
	//
	collName := fmt.Sprintf("following-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Following.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(storageMongo))
	//
	err = s.Create(ctx, model.Source{
		ActorId: "actor0",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name0",
		Summary: "summary0",
	})
	require.Nil(t, err)
	err = s.Create(ctx, model.Source{
		ActorId: "actor1",
		GroupId: "group0",
		UserId:  "user0",
		Type:    "type0",
		Name:    "name1",
		Summary: "summary1",
	})
	require.Nil(t, err)
	err = s.Create(ctx, model.Source{
		ActorId: "actor2",
		GroupId: "group0",
		UserId:  "user1",
		Type:    "type0",
		Name:    "name2",
		Summary: "summary2",
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		filter  model.Filter
		limit   uint32
		cursor  string
		order   model.Order
		pattern string
		page    []string
		err     error
	}{
		"all at once": {
			limit: 10,
			page: []string{
				"actor0",
				"actor1",
				"actor2",
			},
		},
		"all at once w/ user filter": {
			limit: 10,
			filter: model.Filter{
				GroupId: "group0",
				UserId:  "user0",
			},
			page: []string{
				"actor0",
				"actor1",
			},
		},
		"w/ limit": {
			limit: 1,
			page: []string{
				"actor0",
			},
		},
		"w/ cursor": {
			limit:  10,
			cursor: "actor0",
			page: []string{
				"actor1",
				"actor2",
			},
		},
		"w/ cursor desc": {
			limit:  10,
			cursor: "actor1",
			page: []string{
				"actor0",
			},
			order: model.OrderDesc,
		},
		"w/ filter pattern": {
			limit: 10,
			filter: model.Filter{
				Pattern: "mary1",
			},
			page: []string{
				"actor1",
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var page []string
			page, err = s.List(ctx, c.filter, c.limit, c.cursor, c.order)
			assert.Equal(t, c.page, page)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
