package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Http struct {
			Host string `envconfig:"API_HTTP_HOST" required:"true"`
			Port uint16 `envconfig:"API_HTTP_PORT" default:"8080" required:"true"`
		}
		Port   uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
		Writer struct {
			Backoff   time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
			BatchSize uint32        `envconfig:"API_WRITER_BATCH_SIZE" default:"16" required:"true"`
			Uri       string        `envconfig:"API_WRITER_URI" default:"resolver:50051" required:"true"`
		}
		Key struct {
			Public  string `envconfig:"API_KEY_PUBLIC" required:"true"`
			Private string `envconfig:"API_KEY_PRIVATE" required:"true"`
		}
	}
	Db  DbConfig
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
	Search struct {
		Mastodon MastodonConfig
	}
}

type DbConfig struct {
	Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"int-activitypub" required:"true"`
	UserName string `envconfig:"DB_USERNAME" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
	Table    struct {
		Followers struct {
			Name  string `envconfig:"DB_TABLE_NAME_FOLLOWERS" default:"followers" required:"true"`
			Shard bool   `envconfig:"DB_TABLE_SHARD_FOLLOWERS" default:"true"`
		}
		Following struct {
			Cache struct {
				Size int           `envconfig:"DB_TABLE_FOLLOWING_CACHE_SIZE" default:"1024" required:"true"`
				Ttl  time.Duration `envconfig:"DB_TABLE_FOLLOWING_CACHE_TTL" default:"1m" required:"true"`
			}
			Name            string        `envconfig:"DB_TABLE_NAME_FOLLOWING" default:"following" required:"true"`
			Shard           bool          `envconfig:"DB_TABLE_SHARD_FOLLOWING" default:"true"`
			RetentionPeriod time.Duration `envconfig:"DB_TABLE_RETENTION_PERIOD_FOLLOWING" default:"720h" required:"true"`
		}
	}
	Tls struct {
		Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
		Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
	}
}

type MastodonConfig struct {
	Endpoint struct {
		Search string `envconfig:"MASTODON_ENDPOINT_SEARCH" default:"https://mastodon.social/api/v2/search" required:"true"`
		Stream string `envconfig:"MASTODON_ENDPOINT_STREAM" default:"https://streaming.mastodon.social/api/v1/streaming/public?remote=false&only_media=false" required:"true"`
	}
	Client struct {
		Key    string `envconfig:"MASTODON_CLIENT_KEY" required:"true"`
		Secret string `envconfig:"MASTODON_CLIENT_SECRET" required:"true"`
		Token  string `envconfig:"MASTODON_CLIENT_TOKEN" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
