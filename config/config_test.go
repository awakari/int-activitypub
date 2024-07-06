package config

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	os.Setenv("API_HTTP_HOST", "activitypub.awakari.com")
	os.Setenv("API_WRITER_BACKOFF", "23h")
	os.Setenv("API_WRITER_URI", "writer:56789")
	os.Setenv("LOG_LEVEL", "4")
	os.Setenv("API_KEY_PUBLIC", "xxx")
	os.Setenv("API_KEY_PRIVATE", "yyy")
	os.Setenv("DB_TABLE_FOLLOWING_CACHE_SIZE", "1234567")
	os.Setenv("DB_TABLE_FOLLOWING_CACHE_TTL", "89s")
	os.Setenv("API_NODE_NAME", "awakari.com")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, 23*time.Hour, cfg.Api.Writer.Backoff)
	assert.Equal(t, "writer:56789", cfg.Api.Writer.Uri)
	assert.Equal(t, slog.LevelWarn, slog.Level(cfg.Log.Level))
	assert.Equal(t, 1234567, cfg.Db.Table.Following.Cache.Size)
	assert.Equal(t, time.Second*89, cfg.Db.Table.Following.Cache.Ttl)
	assert.Equal(t, time.Hour*720, cfg.Db.Table.Following.RetentionPeriod)
}
