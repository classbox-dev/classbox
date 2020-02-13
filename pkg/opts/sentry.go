package opts

import (
	"github.com/getsentry/sentry-go"
	"log"
)

type Sentry struct {
	Dsn   string `long:"dsn" env:"DSN" description:"sentry dsn" required:"false"`
	Debug bool   `long:"debug" env:"DEBUG" description:"enable debug mode"`
}

func (s *Sentry) Init(env, service string) bool {
	if len(s.Dsn) == 0 {
		return false
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         s.Dsn,
		Debug:       s.Debug,
		Environment: env,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Tags["service"] = service
			return event
		},
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	return true
}
