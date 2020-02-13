package opts

type Docker struct {
	Pull         bool   `long:"pull" env:"PULL" description:"pull course images"`
	Login        bool   `long:"login" env:"LOGIN" description:"log in before pulling the images"`
	Repo         *Repo  `group:"Docker Images Repository" namespace:"repo"  env-namespace:"REPO"`
	BuilderImage string `long:"builder-image" env:"BUILDER_IMAGE" description:"builder image" required:"true"`
	RunnerImage  string `long:"runner-image" env:"RUNNER_IMAGE" description:"runner image" required:"true"`
}

type Repo struct {
	Username string `long:"username" env:"USERNAME" description:"repo username"`
	Password string `long:"password" env:"PASSWORD" description:"repo password"`
	Host     string `long:"host" env:"HOST" description:"repo host"`
}
