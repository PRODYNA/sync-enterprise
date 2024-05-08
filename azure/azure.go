package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msauth "github.com/microsoftgraph/msgraph-sdk-go/msauth"
	"log/slog"
)

type Config struct {
	AzureTenantId     string
	AzureClientId     string
	AzureClientSecret string
	AzureGroup        string
}

type Azure struct {
	config     Config
	azureUsers *AzureUsers
	azclient   *msgraphsdk.GraphServiceClient
}

type AzureUser struct {
	Email       string
	DisplayName string
}

type AzureUsers []AzureUser

func New(ctx context.Context, config Config) (*Azure, error) {
	az := Azure{
		config: config,
	}

	cred, err := azidentity.NewClientSecretCredential(
		config.AzureTenantId,
		config.AzureClientId,
		config.AzureClientSecret,
		nil)
	if err != nil {
		return nil, err
	}

	auth := msauth.(ctx, config.AzureTenantId, config.AzureClientId, config.AzureClientSecret)
	az.azclient, err = msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
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

func (a Azure) Users(ctx context.Context) (AzureUsers, error) {
	if a.azureUsers == nil {
		users := AzureUsers{}

		group, err := a.azclient.Groups().ByGroupId(a.config.AzureGroup).Get(ctx, nil)
		if err != nil {
			return nil, err
		}

		for _, member := range group.GetMembers() {
			slog.Info("Member", slog.Any("member", member))
		}

		a.azureUsers = &users

	}
	return *a.azureUsers, nil
}

func (aus AzureUsers) Contains(email string) bool {
	for _, u := range aus {
		if u.Email == email {
			return true
		}
	}
	return false
}
