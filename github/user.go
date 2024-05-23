package github

import (
	"context"
	"github.com/google/go-github/v61/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log/slog"
)

type Config struct {
	Enterprise string
	Token      string
	DryRun     bool
}

type GitHub struct {
	config       Config
	client       *github.Client
	userlist     GitHubUsers
	enterpriseId string
}

type GitHubUser struct {
	ID    string
	Login string
	Email string
}

type GitHubUsers []GitHubUser

func New(ctx context.Context, config Config) (*GitHub, error) {
	gh := GitHub{
		config: config,
		client: github.NewClient(nil).WithAuthToken(config.Token),
	}

	return &gh, nil
}

func (g *GitHub) Users(ctx context.Context) ([]GitHubUser, error) {
	if g.userlist == nil {
		err := g.loadMembers(ctx)
		if err != nil {
			return nil, err
		}
	}
	return g.userlist, nil
}

func (g GitHub) DryRun() bool {
	return g.config.DryRun
}

func (g *GitHub) loadMembers(ctx context.Context) error {
	slog.Info("Loading members", "enterprise", g.config.Enterprise)
	gitHubUsers := []GitHubUser{}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.config.Token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	var query struct {
		Enterprise struct {
			Id        string
			Slug      string
			Name      string
			OwnerInfo struct {
				SamlIdentityProvider struct {
					ExternalIdentities struct {
						PageInfo struct {
							HasNextPage bool
							EndCursor   githubv4.String
						}
						Edges []struct {
							Node struct {
								User struct {
									ID                      string
									Login                   string
									Name                    string
									ContributionsCollection struct {
										ContributionCalendar struct {
											TotalContributions int
										}
									}
								}
								SamlIdentity struct {
									NameId string
								}
							}
						}
					} `graphql:"externalIdentities(after: $after, first: $first)"`
				}
			}
		} `graphql:"enterprise(slug: $slug)"`
	}

	window := 25
	variables := map[string]interface{}{
		"slug":  githubv4.String(g.config.Enterprise),
		"first": githubv4.Int(window),
		"after": (*githubv4.String)(nil),
	}

	for offset := 0; ; offset += window {
		slog.Debug("Running query", "offset", offset, "window", window)
		err := client.Query(ctx, &query, variables)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to query", "error", err)
			return err
		}

		g.enterpriseId = query.Enterprise.Id

		for _, e := range query.Enterprise.OwnerInfo.SamlIdentityProvider.ExternalIdentities.Edges {
			slog.Debug("GitHub user",
				"id", e.Node.User.ID,
				"login", e.Node.User.Login,
				"email", e.Node.SamlIdentity.NameId)
			u := GitHubUser{
				ID:    e.Node.User.ID,
				Login: e.Node.User.Login,
				Email: e.Node.SamlIdentity.NameId,
			}
			gitHubUsers = append(gitHubUsers, u)
		}

		if !query.Enterprise.OwnerInfo.SamlIdentityProvider.ExternalIdentities.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Enterprise.OwnerInfo.SamlIdentityProvider.ExternalIdentities.PageInfo.EndCursor)
	}

	g.userlist = gitHubUsers

	slog.InfoContext(ctx, "Loaded userlist", "users", len(g.userlist))
	return nil
}

func (g GitHub) EnterpriseId() string {
	return g.enterpriseId
}

func (g GitHub) InviteUser(ctx context.Context, email string, name string) error {
	slog.WarnContext(ctx, "Method not implemented", "email", email, "name", name, "enterprise", g.config.Enterprise)
	return nil
}
