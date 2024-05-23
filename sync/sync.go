package sync

import (
	"context"
	"github.com/prodyna/sync-enterprise/azure"
	"github.com/prodyna/sync-enterprise/github"
	"log/slog"
	"strings"
)

type ActionType int

const (
	// Delete represents a delete action
	Delete ActionType = iota
	Invite ActionType = iota
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
	invite := 0
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

	slog.Info("Checking if Azure is is already in GitHub")
	azureUsers, err := az.Users(ctx)
	if err != nil {
		return err
	}
	for _, azureUser := range azureUsers {
		slog.Debug("Checking user", "email", azureUser.Email, "name", azureUser.DisplayName)
		found := false
		for _, githubUser := range githubUsers {
			if strings.ToLower(githubUser.Email) == strings.ToLower(azureUser.Email) {
				found = true
				break
			}
		}

		if !found {
			slog.Debug("User not in GitHub", "email", azureUser.Email, "name", azureUser.DisplayName)
			action := &Action{
				actionType:  Invite,
				email:       azureUser.Email,
				displayName: azureUser.DisplayName,
			}
			invite++
			actions = append(actions, *action)
		}
	}

	for _, a := range actions {
		switch a.actionType {
		case Invite:
			if gh.DryRun() {
				slog.Info("Dry-run, would invite user",
					"email", a.email,
					"name", a.displayName)
				continue
			} else {
				slog.Info("Inviting user",
					"email", a.email,
					"name", a.displayName)
				err = gh.InviteUser(ctx, a.email, a.displayName)
				if err != nil {
					continue
					// return err
				}
			}
		case Delete:
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
		"invite", invite,
		"stay", stay)

	return nil
}
