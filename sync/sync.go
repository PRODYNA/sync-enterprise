package sync

import (
	"context"
	"github.com/prodyna/delete-from-enterprise/azure"
	"github.com/prodyna/delete-from-enterprise/github"
	"log/slog"
)

type ActionType int

const (
	// nothing represents no action
	Nothing ActionType = iota
	// Delete represents a delete action
	Delete ActionType = iota
)

// Action represents a delete action
type Action struct {
	actionType ActionType
	azureUser  azure.AzureUser
	githubUser github.GitHubUser
}

func Sync(ctx context.Context, az azure.Azure, gh github.GitHub) (err error) {
	slog.Info("Syncing users")
	actions := []Action{}
	delete := 0
	stay := 0

	// load github users
	githubUsers, err := gh.Users(ctx)

	slog.Info("Checking if github users are in Azure group", "count", len(githubUsers), "group", az.Config.AzureGroup)
	for _, githubUser := range githubUsers {
		// check if user is in azure
		inAzure, err := az.IsUserInGroup(ctx, githubUser.Email)
		if err != nil {
			return err
		}

		if !inAzure {
			actions = append(actions, Action{
				actionType: Delete,
				githubUser: githubUser,
			})
			delete++
		} else {
			stay++
		}
	}

	for _, a := range actions {
		if a.actionType == Delete {
			if gh.DryRun() {
				slog.Info("Would delete user", "login", a.githubUser.Login, "email", a.githubUser.Email)
				continue
			}

			slog.Info("Deleting user", "user", "login", a.githubUser.Login, "email", a.githubUser.Email)
			err = gh.DeleteUser(a.githubUser)
			if err != nil {
				return err
			}
		}
	}

	slog.Info("Sync finished", "delete", delete, "leave", stay)

	return nil
}
