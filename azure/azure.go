package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
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
	azclient *msgraphsdk.GraphServiceClient
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
		nil)
	if err != nil {
		return nil, err
	}

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

func (a Azure) IsUserInGroup(ctx context.Context, email string) (bool, error) {
	return false, nil
}
