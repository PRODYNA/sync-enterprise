# sync-enterprise

GitHub Action that syncs GitHub users with Azure AD users.

## Usage on CLI

```
$ Usage of sync-enterprise:
  -azure-client-id string
    	The Azure Client ID.
  -azure-client-secret string
    	The Azure Client Secret.
  -azure-group string
    	The Azure Group.
  -azure-tenant-id string
    	The Azure Tenant ID.
  -dry-run
    	Dry run mode. (default true)
  -github-enterprise string
    	The GitHub Enterprise to query for repositories.
  -github-token string
    	The GitHub Token to use for authentication. 
  ```

## Usage in GitHub Actions

```yaml
name: Sync enterprise

on:
  workflow_dispatch:
  # Every day at 08:00
  schedule:
    - cron: '0 8 * * *'
  push:
    branches:
      - main

jobs:
  sync-enterprise:
    name: Sync enterprise
    runs-on: ubuntu-latest
    steps:

      # Find enterprise users to delete
      - name: Sync enterprise
        uses: prodyna/sync-enterprise@v0.9.2
        with:
          github-token: ${{ secrets.DFE_GITHUB_TOKEN }}
          github-enterprise: "prodyna"
          dry-run: "false"
          azure-group: ${{ vars.DFE_AZURE_GROUP_ID }}
          azure-tenant-id: ${{ vars.DFE_TENANT_ID }}
          azure-client-id: ${{ vars.DFE_AZURE_CLIENT_ID }}
          azure-client-secret: ${{ secrets.DFE_AZURE_CLIENT_SECRET }}
```

## Token permissions

The token needs the following permissions:

* `admin:org`

