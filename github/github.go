package github

import (
	"context"
	"github.com/google/go-github/v61/github"
)

type Config struct {
	Enterprise string
	Token      string
	DryRun     bool
}

type GitHub struct {
	config Config
	client *github.Client
}

type GitHubUser struct {
	Login string
	Email string
}

func New(ctx context.Context, config Config) (*GitHub, error) {
	gh := GitHub{
		config: config,
		client: github.NewClient(nil).WithAuthToken(config.Token),
	}

	return &gh, nil
}

func (g GitHub) Users(ctx context.Context) ([]GitHubUser, error) {
	return nil, nil
}

func (g GitHub) DeleteUser(user GitHubUser) error {
	return nil
}

func (g GitHub) DryRun() bool {
	return g.config.DryRun
}
