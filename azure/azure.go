package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphgocore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"log/slog"
)

type Config struct {
	AzureTenantId     string
	AzureClientId     string
	AzureClientSecret string
	AzureGroup        string
}

type Azure struct {
	Config   Config
	azclient *msgraph.GraphServiceClient
	users    AzureUsers
}

type AzureUser struct {
	Email       string
	DisplayName string
}

type AzureUsers []AzureUser

func New(ctx context.Context, config Config) (*Azure, error) {
	az := Azure{
		Config: config,
	}

	cred, err := azidentity.NewClientSecretCredential(
		config.AzureTenantId,
		config.AzureClientId,
		config.AzureClientSecret,
		&azidentity.ClientSecretCredentialOptions{})
	if err != nil {
		return nil, err
	}

	az.azclient, err = msgraph.NewGraphServiceClientWithCredentials(cred, []string{})
	if err != nil {
		return nil, err
	}

	// try to connect to the group
	group, err := az.azclient.Groups().ByGroupId(config.AzureGroup).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	slog.Info("Connected to group", "group", *group.GetDisplayName())

	return &az, nil
}

func (az *Azure) Users(ctx context.Context) ([]AzureUser, error) {
	if az.users == nil {
		users := []AzureUser{}

		top := int32(999)
		query := groups.ItemMembersGraphUserRequestBuilderGetQueryParameters{
			Select: []string{"id", "displayName", "mail"},
			Top:    &top,
		}

		options := &groups.ItemMembersGraphUserRequestBuilderGetRequestConfiguration{
			QueryParameters: &query,
		}

		result, err := az.azclient.Groups().ByGroupId(az.Config.AzureGroup).Members().GraphUser().Get(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("error getting group members: %w", err)
		}
		slog.Info("result", slog.Any("result", result))

		pageIterator, err := msgraphgocore.NewPageIterator[*models.User](result, az.azclient.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
		if err != nil {
			return nil, fmt.Errorf("error creating page iterator: %w", err)
		}

		err = pageIterator.Iterate(ctx, func(user *models.User) bool {
			if user != nil {
				slog.Info("Azure group member",
					"email", *user.GetMail(),
					"displayName", *user.GetDisplayName())
				users = append(users, AzureUser{
					Email:       *user.GetMail(),
					DisplayName: *user.GetDisplayName(),
				})
			}
			return true
		})

		az.users = users
	}

	return az.users, nil
}

func (az *Azure) IsUserInGroup(ctx context.Context, email string) (bool, error) {
	users, err := az.Users(ctx)
	if err != nil {
		return false, err
	}

	for _, user := range users {
		if user.Email == email {
			return true, nil
		}
	}

	return false, nil
}
