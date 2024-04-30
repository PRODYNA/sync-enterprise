package main

import (
	"github.com/prodyna/delete-from-enterprise/config"
	"log/slog"
	"os"
)

func main() {
	c, err := config.New()
	if err != nil {
		slog.Error("Unable to create config", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration",
		"githubEnterprise", c.GitHub.Enterprise,
		"githubToken", "***",
		"azureClientId", c.Azure.ClientId,
		"azureClientSecret", "***",
		"azureTenantId", c.Azure.TenantId,
		"azureGroup", c.Azure.Group,
		"dryRun", c.DryRun)
}
