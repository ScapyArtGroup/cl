package list

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/ghinstance"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmd/secret/shared"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/utils"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)

	OrgName string
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Long:  "List secrets for a repository or organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo

			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.OrgName, "org", "", "List secrets for an organization")
	cmd.Flags().Lookup("org").NoOptDefVal = "@owner"

	return cmd
}

func listRun(opts *ListOptions) error {
	c, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}
	client := api.NewClientFromHTTP(c)

	var baseRepo ghrepo.Interface
	if opts.OrgName == "" || opts.OrgName == "@owner" {
		baseRepo, err = opts.BaseRepo()
		if err != nil {
			return fmt.Errorf("could not determine base repo: %w", err)
		}
	}

	orgName := opts.OrgName
	host := ghinstance.OverridableDefault()
	if orgName == "@owner" {
		orgName = baseRepo.RepoOwner()
		host = baseRepo.RepoHost()
	}

	var secrets []Secret
	if orgName != "" {
		secrets, err = getOrgSecrets(client, host, orgName)
	} else {
		secrets, err = getRepoSecrets(client, baseRepo)
	}

	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	tp := utils.NewTablePrinter(opts.IO)
	for _, secret := range secrets {
		tp.AddField(secret.Name, nil, nil)
		updatedAt := secret.UpdatedAt.Format("2006-01-02")
		if opts.IO.IsStdoutTTY() {
			updatedAt = fmt.Sprintf("Updated %s", updatedAt)
		}
		tp.AddField(updatedAt, nil, nil)
		if secret.Visibility != "" {
			if opts.IO.IsStdoutTTY() {
				tp.AddField(fmtVisibility(secret), nil, nil)
			} else {
				tp.AddField(strings.ToUpper(secret.Visibility), nil, nil)
			}
		}
		tp.EndRow()
	}

	err = tp.Render()
	if err != nil {
		return err
	}

	return nil
}

type Secret struct {
	Name       string
	UpdatedAt  time.Time `json:"updated_at"`
	Visibility string
}

func fmtVisibility(s Secret) string {
	switch s.Visibility {
	case shared.VisAll:
		return "Visible to all repositories"
	case shared.VisPrivate:
		return "Visible to private repositories"
	case shared.VisSelected:
		// TODO print how many? print which ones?
		return "Visible to selected repositories"
	}
	return ""
}

func getOrgSecrets(client *api.Client, host, orgName string) ([]Secret, error) {
	return getSecrets(client, host, fmt.Sprintf("orgs/%s/actions/secrets", orgName))
}

func getRepoSecrets(client *api.Client, repo ghrepo.Interface) ([]Secret, error) {
	return getSecrets(client, repo.RepoHost(), fmt.Sprintf("repos/%s/actions/secrets",
		ghrepo.FullName(repo)))
}

type secretsPayload struct {
	Secrets []Secret
}

func getSecrets(client *api.Client, host, path string) ([]Secret, error) {
	result := secretsPayload{}

	err := client.REST(host, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return result.Secrets, nil
}
