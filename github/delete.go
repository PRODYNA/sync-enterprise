package github

import (
	"context"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log/slog"
)

func (g GitHub) DeleteUser(ctx context.Context, userId string) error {
	enterpriseId := g.enterpriseId
	slog.Info("Deleting user", "userId", userId, "enterprise", g.config.Enterprise, "enterpriseId", g.enterpriseId)

	var mutation struct {
		RemoveEnterpriseMember struct {
			ClientMutationId string
			Enterprise       struct {
				ID string
			}
			User struct {
				ID string
			}
			Viewer struct {
				ID string
			}
		} `graphql:"removeEnterpriseMember(input:$input)"`
	}

	input := githubv4.RemoveEnterpriseMemberInput{
		EnterpriseID: githubv4.ID(enterpriseId),
		UserID:       githubv4.ID(userId),
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.config.Token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	err := client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		slog.Warn("Unable to delete user", "userId", userId, "enterprise", g.config.Enterprise, "error", err)
		return nil
	}
	slog.Info("User deleted", "userId", userId, "enterprise", g.config.Enterprise)

	return nil
}
