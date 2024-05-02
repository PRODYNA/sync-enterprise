package sync

import (
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

func Sync(az azure.Azure, gh github.GitHub) (err error) {
	actions := []Action{}

	githubUsers, err := gh.Users()
	azureUsers, err := az.Users()
	if err != nil {
		return err
	}
	for _, githubUser := range githubUsers {
		if azureUsers.Contains(githubUser.Email) {
			actions = append(actions, Action{
				actionType: Nothing,
				azureUser:  azure.AzureUser(githubUser.Email),
				githubUser: githubUser,
			})
		} else {
			actions = append(actions, Action{
				actionType: Delete,
				azureUser:  azure.AzureUser(githubUser.Email),
				githubUser: githubUser,
			})
		}
	}

	for _, a := range actions {
		if a.actionType == Delete {
			if gh.DryRun() {
				slog.Info("Would delete user", "user", a.githubUser)
				continue
			}

			slog.Info("Deleting user", "user", a.githubUser)
			err = gh.DeleteUser(a.githubUser)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
