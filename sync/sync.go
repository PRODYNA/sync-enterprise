package sync

import (
	"context"
	"github.com/prodyna/delete-from-enterprise/azure"
	"github.com/prodyna/delete-from-enterprise/github"
	"log/slog"
)

type ActionType int

const (
	// Delete represents a delete action
	Delete ActionType = iota
)

// Action represents a delete action
type Action struct {
	actionType  ActionType
	displayName string
	email       string
	login       string
	id          string
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
		slog.Debug("Checking user", "login", githubUser.Login, "email", githubUser.Email)
		// check if user is in azure
		inAzure, name, err := az.IsUserInGroup(ctx, githubUser.Email)
		if err != nil {
			return err
		}

		if !inAzure {
			slog.Debug("User not in Azure", "login", githubUser.Login, "email", githubUser.Email)
			action := &Action{
				actionType: Delete,
				id:         githubUser.ID,
				email:      githubUser.Email,
				login:      githubUser.Login,
			}
			if name != nil {
				action.displayName = *name
			}
			actions = append(actions, *action)
			delete++
		} else {
			slog.Debug("User in Azure", "login", githubUser.Login, "email", githubUser.Email, "name", *name)
			stay++
		}
	}

	for _, a := range actions {
		if a.actionType == Delete {
			if gh.DryRun() {
				slog.Info("Dry-run, would delete user",
					"login", a.login,
					"email", a.email,
					"name", a.displayName)
				continue
			}

			slog.Info("Deleting user",
				"login", a.login,
				"userId", a.id,
				"email", a.email,
				"name", a.displayName)
			err = gh.DeleteUser(ctx, a.id)
			if err != nil {
				continue
				// return err
			}
		}
	}

	slog.Info("Sync finished",
		"delete", delete,
		"stay", stay)

	return nil
}
