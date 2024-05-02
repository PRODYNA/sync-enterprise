package config

import (
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"
	"strconv"
)

const (
	keyGithubEnterprise  = "github-enterprise"
	keyGithubToken       = "github-token"
	keyAzureClientId     = "azure-client-id"
	keyAzureClientSecret = "azure-client-secret"
	keyAzureTenantId     = "azure-tenant-id"
	keyAzureGroup        = "azure-group"
	keyDryRun            = "dry-run"

	keyGitHubEnterpriseEnvironment  = "GITHUB_ENTERPRISE"
	keyGitHubTokenEnvironment       = "GITHUB_TOKEN"
	keyAzureClientIdEnvironment     = "AZURE_CLIENT_ID"
	keyAzureClientSecretEnvironment = "AZURE_CLIENT_SECRET"
	keyAzureTenantIdEnvironment     = "AZURE_TENANT_ID"
	keyAzureGroupEnvironment        = "AZURE_GROUP"
	keyDryRunEnvironment            = "DRY_RUN"
)

type GitHub struct {
	Enterprise string
	Token      string
}

type Azure struct {
	ClientId     string
	ClientSecret string
	TenantId     string
	Group        string
}

type Config struct {
	GitHub GitHub
	Azure  Azure
	DryRun bool
}

func New() (*Config, error) {
	c := Config{}
	flag.StringVar(&c.GitHub.Token, keyGithubToken, lookupEnvOrString(keyGitHubTokenEnvironment, ""), "The GitHub Token to use for authentication.")
	flag.StringVar(&c.GitHub.Enterprise, keyGithubEnterprise, lookupEnvOrString(keyGitHubEnterpriseEnvironment, ""), "The GitHub Enterprise to query for repositories.")
	flag.StringVar(&c.Azure.ClientId, keyAzureClientId, lookupEnvOrString(keyAzureClientIdEnvironment, ""), "The Azure Client ID.")
	flag.StringVar(&c.Azure.ClientSecret, keyAzureClientSecret, lookupEnvOrString(keyAzureClientSecretEnvironment, ""), "The Azure Client Secret.")
	flag.StringVar(&c.Azure.TenantId, keyAzureTenantId, lookupEnvOrString(keyAzureTenantIdEnvironment, ""), "The Azure Tenant ID.")
	flag.StringVar(&c.Azure.Group, keyAzureGroup, lookupEnvOrString(keyAzureGroupEnvironment, ""), "The Azure Group.")
	flag.BoolVar(&c.DryRun, keyDryRun, lookupEnvOrBool(keyDryRunEnvironment, false), "Dry run mode.")

	flag.Parse()

	if c.GitHub.Token == "" {
		slog.Error("GitHub Token is required")
		return nil, errors.New("GitHub Token is required")
	}
	if c.GitHub.Enterprise == "" {
		slog.Error("GitHub Enterprise is required")
		return nil, errors.New("GitHub Enterprise is required")
	}
	if c.Azure.ClientId == "" {
		slog.Error("Azure Client ID is required")
		return nil, errors.New("Azure Client ID is required")
	}
	if c.Azure.ClientSecret == "" {
		slog.Error("Azure Client Secret is required")
		return nil, errors.New("Azure Client Secret is required")
	}
	if c.Azure.TenantId == "" {
		slog.Error("Azure Tenant ID is required")
		return nil, errors.New("Azure Tenant ID is required")
	}
	if c.Azure.Group == "" {
		slog.Error("Azure Group is required")
		return nil, errors.New("Azure Group is required")
	}

	return &c, nil
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("LookupEnvOrBool[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
