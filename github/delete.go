package github

import (
	"context"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log/slog"
)

func (g GitHub) DeleteUser(ctx context.Context, login string) error {
	slog.Info("Deleting user", "login", login, "enterprise", g.config.Enterprise)

	var mutation struct {
		RemoveEnterpriseMember struct {
			Input struct {
				ClientMutationId string
				EnterpriseId     string
				userId           string
			}
		} `graphql:"removeEnterpriseMember(input: $input)"`
	}

	input := map[string]interface{}{
		"input": map[string]interface{}{
			"clientMutationId": "delete-user",
			"enterpriseId":     g.config.Enterprise,
			"userId":           login,
		},
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.config.Token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	err := client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		slog.Warn("Unable to delete user", "login", login, "enterprise", g.config.Enterprise, "error", err)
		return err
	}
	slog.Info("User deleted", "login", login, "enterprise", g.config.Enterprise)

	return nil
}
