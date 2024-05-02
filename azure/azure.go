package azure

type Config struct {
	AzureTenantId     string
	AzureClientId     string
	AzureClientSecret string
	AzureGroup        string
}

type Azure struct {
	config     Config
	azureUsers AzureUsers
}

type AzureUser string

type AzureUsers []AzureUser

func New(config Config) (*Azure, error) {
	return &Azure{
		config: config,
	}, nil
}

func (a Azure) Users() (AzureUsers, error) {
	return a.azureUsers, nil
}

func (aus AzureUsers) Contains(email string) bool {
	for _, u := range aus {
		if string(u) == email {
			return true
		}
	}
	return false
}
