package original

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"context"

	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/joho/godotenv"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	users "github.com/microsoftgraph/msgraph-sdk-go/users"
	// log "github.com/sirupsen/logrus"
)

const (
	envFile = ".env"
)

type Licenses struct {
	Users []User `json:"users"`
}

type User struct {
	GitHubLogin      string `json:"github_com_login"`
	GitHubName       string `json:"github_com_name"`
	GitHubSAMLNameID string `json:"github_com_saml_name_id"`
}

type Repository struct {
	FullName   string `json:"full_name"`
	Private    bool   `json:"private"`
	Disabled   bool   `json:"disabled"`
	Visibility string `json:"visibility"`
}

type RepositoryResponse []Repository

type RepositoryWithContributors struct {
	Repo        Repository
	Contributor []Contributor
}

type Organization struct {
	Login       string `json:"login"`
	Description string `json:"description"`
}

type Plan struct {
	Name string `json:"name"`
}

type Contributor struct {
	Login    string `json:"login"`
	RoleName string `json:"role_name"`
}

type ContributorResponse []Contributor

type EnterpriseOrganisationsResponse struct {
	Data struct {
		Enterprise struct {
			Organizations struct {
				Edges []struct {
					Node struct {
						Login       string `json:"login"`
						Description string `json:"description"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"organizations"`
		} `json:"enterprise"`
	} `json:"data"`
}

func InitLogging() {
	level := new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
	level.Set(slog.LevelDebug)
}

func main() {

	InitLogging()

	slog.Debug("Starting application")

	slog.Info("Loading environment variables", slog.String("file", envFile))
	err := godotenv.Load(envFile)
	if err != nil {
		slog.Warn("No environment variables file found", slog.String("file", envFile))
	}

	licencedUsers, err := getGitHubLicencedUsers()
	if err != nil {
		slog.Error("Failed to get Github Enterprise users", slog.Any("error", err))
		os.Exit(1)
	}

	orgs, err := getGithubEnterpriseOrgs()
	if err != nil {
		slog.Error("Failed to get Github Organisations that belong to the Enterprise", slog.Any("error", err))
		os.Exit(1)
	}

	repos, err := getGithubPrivateReposForOrgs(orgs)
	if err != nil {
		slog.Error("Failed to get Github private repos", slog.Any("error", err))
		os.Exit(1)
	}

	reposWithContributors, err := getGithubReposWithContributors(repos)
	if err != nil {
		slog.Error("Failed to get contributors from the private repos", slog.Any("error", err))
		os.Exit(1)
	}

	// Filter out users without SAML name ID
	prodynaUsers := []User{}
	outsideUsers := []User{}
	removeUsers := []User{}

	slog.Debug("Iterating all licenced users", slog.Int("count", len(licencedUsers)))
	for _, user := range licencedUsers {
		if user.GitHubSAMLNameID != "" {
			prodynaUsers = append(prodynaUsers, user)
		} else {
			outsideUsers = append(outsideUsers, user)
		}
	}

	// Use Service Principal
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		slog.Error("Failed to obtain a credential", slog.Any("error", err))
		os.Exit(1)
	}

	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		slog.Error("Failed to create a client", slog.Any("error", err))
		os.Exit(1)
	}
	// "User.Read.All"

	// Check if users are active in Azure AD
	slog.Debug("Iterating all users", slog.Int("count", len(prodynaUsers)))
	for _, user := range prodynaUsers {
		active, err := isAzureAdUserActive(user.GitHubSAMLNameID, client)
		if err != nil {
			slog.Error("Failed to check if user is active in Azure AD", slog.String("user", user.GitHubSAMLNameID))
			continue
		}
		if !active {
			slog.Debug("User is not active in Azure AD", slog.String("user", user.GitHubSAMLNameID))
			removeUsers = append(removeUsers, user)
		} else {
			slog.Debug("User is active in Azure AD", slog.String("user", user.GitHubSAMLNameID))
		}
	}

	slog.Debug("Iterating all users to remove", slog.Int("count", len(removeUsers)))
	for _, user := range removeUsers {
		slog.Debug("Remove user from Github Enterprise", slog.String("samlName", user.GitHubSAMLNameID), slog.String("githubLogin", user.GitHubLogin), slog.String("githubName", user.GitHubName))
	}

	slog.Debug("Iterating all outside collaborators", slog.Int("count", len(outsideUsers)))
	for _, user := range outsideUsers {
		for _, repo := range reposWithContributors {
			if isContributor(user, repo) {
				slog.Warn("Review outside collaborator access", slog.String("githubLogin", user.GitHubLogin), slog.String("repoFullName", repo.Repo.FullName))
			}
		}
	}
}

// Check if a given user is contributor to a given repository
func isContributor(user User, repo RepositoryWithContributors) bool {
	for _, contributor := range repo.Contributor {
		if contributor.Login == user.GitHubLogin {
			return true
		}
	}
	return false
}

// Function to get a Github API client
func getGithubClient() *http.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return tc
}

// Function to check if user is active in Azure AD
func isAzureAdUserActive(userPrincipalName string, client *msgraphsdk.GraphServiceClient) (bool, error) {
	// Build the filter to search for the user by their email address.
	filter := "mail eq '" + userPrincipalName + "'"

	// Build the request parameters to select the userPrincipalName and accountEnabled
	// properties and to use the filter.
	requestParameters := users.UsersRequestBuilderGetQueryParameters{
		Filter: &filter,
		Select: []string{"userPrincipalName", "accountEnabled"},
	}

	// Build the request configuration to use the request parameters.
	configuration := users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &requestParameters,
	}

	// Send the request to Azure AD.
	result, err := client.Users().Get(context.Background(), &configuration)
	if err != nil {
		return false, err
	}

	// If the user was found, return their accountEnabled property.
	value := result.GetValue()
	if len(value) == 0 {
		return false, nil
	}
	accountEnabled := value[0].GetAccountEnabled()
	return *accountEnabled, nil
}

// Function to read Github Enterprise users from API
func getGitHubLicencedUsers() ([]User, error) {
	slog.Debug("Reading Github Enterprise licenced users")

	licencedUsers := []User{}

	// Repeat request until all pages are read
	tc := getGithubClient()
	response, err := tc.Get("https://api.github.com/enterprises/prodyna/consumed-licenses?per_page=100")
	nextLink := "first"
	for nextLink != "" {
		if err != nil {
			slog.Error("API call to Github failed", slog.Any("error", err))
			return nil, err
		}

		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Unable to read response", slog.Any("error", err))
			return nil, err
		}

		var licenses Licenses
		err = json.Unmarshal([]byte(responseData), &licenses)
		if err != nil {
			slog.Error("Unable to parse Github Enterprise users", slog.Any("error", err))
			return nil, err
		}
		licencedUsers = append(licencedUsers, licenses.Users...)

		nextLink = getNextLink(response.Header.Get("Link"))
		if nextLink != "" {
			response, err = tc.Get(nextLink)
			if err != nil {
				slog.Error("Failed to get next page of Github Enterprise users", slog.Any("error", err))
				return nil, err
			}
		}
	}
	return licencedUsers, nil
}

// Function to get all repos from a Github Organisation
func getGithubPrivateReposForOrgs(orgs []Organization) ([]Repository, error) {
	slog.Debug("Filtering for Github private repos")

	repos := []Repository{}
	for _, org := range orgs {
		orgRepos, err := getGitRepos(org)
		if err != nil {
			return nil, err
		}
		for _, orgRepo := range orgRepos {
			if orgRepo.Private {
				repos = append(repos, orgRepo)
			}
		}
	}
	return repos, nil
}

func getGitRepos(org Organization) ([]Repository, error) {
	slog.Debug("Reading Github repos", slog.String("organisation", org.Login))

	repos := []Repository{}

	tc := getGithubClient()
	response, err := tc.Get("https://api.github.com/orgs/" + org.Login + "/repos?per_page=100")
	nextLink := "first"
	for nextLink != "" {
		if err != nil {
			slog.Error("API call to Github failed", slog.Any("error", err))
			return nil, err
		}

		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Unable to read response", slog.Any("error", err))
			return nil, err
		}

		var repoPage RepositoryResponse
		err = json.Unmarshal([]byte(responseData), &repoPage)
		if err != nil {
			slog.Error("Unable to parse organisation repos", slog.Any("error", err))
			return nil, err
		}

		repos = append(repos, repoPage...)

		nextLink = getNextLink(response.Header.Get("Link"))
		if nextLink != "" {
			response, err = tc.Get(nextLink)
			if err != nil {
				slog.Error("Failed to get next page of Github Organisation repos", slog.Any("error", err))
				return nil, err
			}
		}
	}
	return repos, nil
}

func getGithubReposWithContributors(repos []Repository) ([]RepositoryWithContributors, error) {
	slog.Debug("Enriching Github repos with contributors")

	reposWithContributors := []RepositoryWithContributors{}
	for _, repo := range repos {
		contributors, err := getGithubRepoContributors(repo)
		if err != nil {
			return nil, err
		}
		reposWithContributors = append(reposWithContributors, RepositoryWithContributors{repo, contributors})
	}

	return reposWithContributors, nil
}

func getGithubRepoContributors(repo Repository) ([]Contributor, error) {
	slog.Debug("Reading Github repo contributors", slog.String("repository", repo.FullName))

	contributors := []Contributor{}

	tc := getGithubClient()
	response, err := tc.Get("https://api.github.com/repos/" + repo.FullName + "/collaborators?per_page=100")
	nextLink := "first"
	for nextLink != "" {
		if err != nil {
			slog.Error("API call to Github failed", slog.Any("error", err))
			return nil, err
		}

		if response.StatusCode != 200 {
			slog.Error("Unable to read repo contributors", slog.String("repository", repo.FullName), slog.Int("status code", response.StatusCode))
			return nil, err
		}
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Unable to read response", slog.Any("error", err))
			return nil, err
		}

		var contributorsPage ContributorResponse
		err = json.Unmarshal([]byte(responseData), &contributorsPage)
		if err != nil {
			slog.Error("Unable to parse repo contributors", slog.Any("error", err), slog.String("response", string(responseData)))
			return nil, err
		}
		contributors = append(contributors, contributorsPage...)

		nextLink = getNextLink(response.Header.Get("Link"))
		if nextLink != "" {
			response, err = tc.Get(nextLink)
			if err != nil {
				slog.Error("Failed to get next page of repo contributors", slog.Any("error", err))
				return nil, err
			}
		}
	}
	return contributors, nil
}

func getGithubEnterpriseOrgs() ([]Organization, error) {
	slog.Debug("Reading Github Enterprise organisations")

	orgs := []Organization{}

	jsonStr, err := json.Marshal(map[string]string{"query": `query{enterprise(slug: "prodyna") {organizations(first: 100) {edges {node {login}}}}}`})
	if err != nil {
		slog.Error("Failed to marshal JSON", slog.Any("error", err))
		return nil, err
	}

	tc := getGithubClient()
	resp, err := tc.Post("https://api.github.com/graphql", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		slog.Error("API call to Github failed", slog.Any("error", err))
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Unable to read response", slog.Any("error", err))
		return nil, err
	}

	var data EnterpriseOrganisationsResponse
	json.Unmarshal(body, &data)
	for _, org := range data.Data.Enterprise.Organizations.Edges {
		orgs = append(orgs, Organization{org.Node.Login, org.Node.Description})
	}

	return orgs, nil
}

// Function to extract the next link from the Link header
func getNextLink(linkHeader string) string {
	navLinks := strings.Split(linkHeader, ",")
	for _, navLink := range navLinks {
		if strings.Contains(navLink, "rel=\"next\"") {
			navLink = strings.Trim(navLink, " ")
			return navLink[1:strings.Index(navLink, ">")]
		}
	}
	return ""
}
