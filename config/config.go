package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api ApiConfig
	Db  DbConfig
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
}

type ApiConfig struct {
	Http struct {
		Host string `envconfig:"API_HTTP_HOST" required:"true"`
		Port uint16 `envconfig:"API_HTTP_PORT" default:"8080" required:"true"`
	}
	Port      uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
	EventType string `envconfig:"API_EVENT_TYPE" required:"true" default:"com_awakari_activitypub_v1"`
	Interests struct {
		Uri              string `envconfig:"API_INTERESTS_URI" required:"true" default:"subscriptions-proxy:50051"`
		DetailsUriPrefix string `envconfig:"API_INTERESTS_DETAILS_URI_PREFIX" required:"true" default:"https://awakari.com/sub-details.html?id="`
	}
	Reader ReaderConfig
	Writer struct {
		Backoff   time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
		BatchSize uint32        `envconfig:"API_WRITER_BATCH_SIZE" default:"16" required:"true"`
		Uri       string        `envconfig:"API_WRITER_URI" default:"resolver:50051" required:"true"`
	}
	Actor struct {
		Name string `envconfig:"API_ACTOR_NAME" required:"true" default:"awakari"`
		Type string `envconfig:"API_ACTOR_TYPE" required:"true" default:"Service"`
	}
	Key struct {
		Public  string `envconfig:"API_KEY_PUBLIC" required:"true"`
		Private string `envconfig:"API_KEY_PRIVATE" required:"true"`
	}
	Node struct {
		Name        string `envconfig:"API_NODE_NAME" required:"true"`
		Description string `envconfig:"API_NODE_DESCRIPTION" required:"true" default:"Awakari Fediverse Integration"`
	}
	Prometheus PrometheusConfig
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

type PrometheusConfig struct {
	Uri string `envconfig:"API_PROMETHEUS_URI" default:"http://prometheus-server:80" required:"true"`
}

type ReaderConfig struct {
	Uri      string `envconfig:"API_READER_URI" default:"http://reader:8080/v1" required:"true"`
	CallBack struct {
		Protocol string `envconfig:"API_READER_CALLBACK_PROTOCOL" default:"http" required:"true"`
		Host     string `envconfig:"API_READER_CALLBACK_HOST" default:"int-activitypub" required:"true"`
		Port     uint16 `envconfig:"API_READER_CALLBACK_PORT" default:"8081" required:"true"`
		Path     string `envconfig:"API_READER_CALLBACK_PATH" default:"/v1/callback" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
