package main

import (
	"context"
	"github.com/prodyna/sync-enterprise/azure"
	"github.com/prodyna/sync-enterprise/config"
	"github.com/prodyna/sync-enterprise/github"
	"github.com/prodyna/sync-enterprise/meta"
	"github.com/prodyna/sync-enterprise/sync"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.SetDefault(logger)
	c, err := config.New()
	if err != nil {
		slog.Error("Unable to create config", "error", err)
		os.Exit(1)
	}
	slog.Info("Starting Sync Enterprise", "version", meta.Version)
	slog.Info("Configuration",
		"githubEnterprise", c.GitHub.Enterprise,
		"githubToken", "***",
		"azureClientId", c.Azure.ClientId,
		"azureClientSecret", "***",
		"azureTenantId", c.Azure.TenantId,
		"azureGroup", c.Azure.Group,
		"dryRun", c.DryRun)

	az, err := azure.New(ctx, azure.Config{
		AzureClientId:     c.Azure.ClientId,
		AzureClientSecret: c.Azure.ClientSecret,
		AzureTenantId:     c.Azure.TenantId,
		AzureGroup:        c.Azure.Group,
	})
	if err != nil {
		slog.Error("Unable to create Azure client", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to azure",
		"tenantId", c.Azure.TenantId,
		"clientId", c.Azure.ClientId,
		"group", c.Azure.Group)

	gh, nil := github.New(ctx, github.Config{
		Enterprise: c.GitHub.Enterprise,
		Token:      c.GitHub.Token,
		DryRun:     c.DryRun,
	})
	if err != nil {
		slog.Error("Unable to create GitHub client", "error", err)
		os.Exit(1)

	}
	slog.Info("Connected to GitHub",
		"enterprise", c.GitHub.Enterprise,
		"token", "***")

	err = sync.Sync(ctx, *az, *gh)
	if err != nil {
		slog.Error("Unable to sync", "error", err)
		os.Exit(1)
	}

}
