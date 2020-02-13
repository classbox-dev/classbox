package opts

// DB contains database settings
type Env struct {
	Type string `long:"type" env:"TYPE" description:"environment type" choice:"development" choice:"testing" choice:"production" required:"true" default:"development"`
}
