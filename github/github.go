package github

type Config struct {
	Enterprise string
	Token      string
	DryRun     bool
}

type GitHub struct {
	config Config
}

type GitHubUser struct {
	Login string
	Email string
}

func New(config Config) (*GitHub, error) {
	return &GitHub{
		config: config,
	}, nil
}

func (g GitHub) Users() ([]GitHubUser, error) {
	return nil, nil
}

func (g GitHub) DeleteUser(user GitHubUser) error {
	return nil
}

func (g GitHub) DryRun() bool {
	return g.config.DryRun
}
