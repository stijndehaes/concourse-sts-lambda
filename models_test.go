package handler_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	handler "github.com/telia-oss/concourse-sts-lambda"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    handler.Team
	}{
		{
			description: "Unmarshal works as intended",
			input: strings.TrimSpace(`
{
    "name": "team",
    "accounts": [{
	    "name": "account",
	    "roleArn": "role"
	}]
}
`),
			expected: handler.Team{
				Name: "team",
				Accounts: []*handler.Account{
					{
						Name:     "account",
						RoleArn:  "role",
						Duration: 3600,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var output handler.Team
			err := json.Unmarshal([]byte(tc.input), &output)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !reflect.DeepEqual(output, tc.expected) {
				got, err := json.Marshal(output)
				if err != nil {
					t.Fatalf("failed to marshal output: %s", err)
				}
				want, err := json.Marshal(tc.expected)
				if err != nil {
					t.Fatalf("failed to marshal expected: %s", err)
				}
				t.Errorf("\ngot:\n%s\nwant:\n%s\n", got, want)
			}
		})
	}
}

func TestAccount_UnmarshalJSON(t *testing.T) {
	accountWithNormalDuration := `
{
	"name": "account",
	"roleArn": "role",
	"duration": 3600
}`
	var output handler.Account
	err := json.Unmarshal([]byte(accountWithNormalDuration), &output)
	if err != nil {
		t.Fatalf("expected not to fail because duration is 3600: %s", err)
	}

}

func TestAccount_UnmarshalJSON_WithoutDuration(t *testing.T) {
	var output handler.Account
	accountWithoutDuration := `
{
	"name": "account",
	"roleArn": "role"
}`
	err := json.Unmarshal([]byte(accountWithoutDuration), &output)
	if err != nil {
		t.Fatalf("expected not to fail because duration is not filled in: %s", err)
	}
	if output.Duration != 3600 {
		t.Fatalf("Default duration should be 3600 but is: %d", output.Duration)
	}
}

func TestSecretPath(t *testing.T) {
	tests := []struct {
		description string
		template    string
		team        string
		account     string
		expected    string
		shouldError bool
	}{
		{
			description: "template works as intended",
			template:    "/concourse/{{.Team}}/{{.Account}}",
			team:        "TEAM",
			account:     "ACCOUNT",
			expected:    "/concourse/TEAM/ACCOUNT",
			shouldError: false,
		},
		{
			description: "fails if the template expects more parameters",
			template:    "/concourse/{{.Team}}/{{.Account}}/{{.Something}}",
			team:        "TEAM",
			account:     "ACCOUNT",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got, err := handler.NewSecretPath(tc.team, tc.account, tc.template).String()

			if tc.shouldError && err == nil {
				t.Fatal("expected an error to occur")
			}

			if !tc.shouldError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if want := tc.expected; got != want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
			}
		})
	}
}
