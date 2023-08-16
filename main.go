package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	pagerdutyProvider "github.com/cdktf/cdktf-provider-pagerduty-go/pagerduty/v9/provider"
	"github.com/cdktf/cdktf-provider-pagerduty-go/pagerduty/v9/team"
	"github.com/cdktf/cdktf-provider-pagerduty-go/pagerduty/v9/teammembership"
	"github.com/cdktf/cdktf-provider-pagerduty-go/pagerduty/v9/user"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

func NewMyStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	token := os.Getenv("PAGERDUTY_TOKEN")

	pagerdutyProvider.NewPagerdutyProvider(stack, jsii.String("pagerduty"), &pagerdutyProvider.PagerdutyProviderConfig{
		Token: jsii.String(token),
	})

	f, err := os.Open("users.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	usersCSVData, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	usersData := generateUserData(usersCSVData)

	users := generateUsers(stack, usersData)
	teams := generateTeams(stack, usersData)
	generateTeamMemberships(stack, usersData, users, teams)

	// cdktf.NewTerraformOutput(stack, jsii.String("user_local_1_config"), &cdktf.TerraformOutputConfig{
	// 	Value: pagerDutyLocalUser1,
	// })

	return stack
}

type UserRecord struct {
	Key         string
	Name        string
	Email       string
	Role        string
	JobTitle    string
	CountryCode string
	Phone       string
	Sms         string
	Team        string
	// TODO: Add Team Role for team memberships to work correctly.
}

func generateUserData(csvData [][]string) []UserRecord {
	var usersData []UserRecord
	for i, line := range csvData {
		if i > 0 { // omit header line
			var rec UserRecord
			for j, field := range line {
				if j == 0 {
					rec.Name = field
				} else if j == 1 {
					rec.Email = field
				}
				switch j {
				case 0:
					rec.Key = field
				case 1:
					rec.Name = field
				case 2:
					rec.Email = field
				case 3:
					rec.Role = field
				case 4:
					rec.JobTitle = field
				case 5:
					rec.CountryCode = field
				case 6:
					rec.Phone = field
				case 7:
					rec.Sms = field
				case 8:
					rec.Team = field
				default:
					fmt.Printf("j is out of range for value %d", j)
				}
			}
			usersData = append(usersData, rec)
		}
	}

	return usersData
}

func generateUsers(scope constructs.Construct, usersData []UserRecord) map[string]user.User {
	users := make(map[string]user.User)
	for _, u := range usersData {
		users[u.Email] = user.NewUser(scope, jsii.String(u.Key), &user.UserConfig{
			Name:     jsii.String(u.Name),
			Email:    jsii.String(u.Email),
			Role:     jsii.String(u.Role),
			JobTitle: jsii.String(u.JobTitle),
		})
	}

	return users
}

func generateTeams(scope constructs.Construct, usersCSVData []UserRecord) map[string]team.Team {
	teams := make(map[string]team.Team)

	teamNames := []string{}
	for _, u := range usersCSVData {
		teamNames = append(teamNames, u.Team)
	}
	teamNames = unique(teamNames)

	for _, name := range teamNames {
		teams[name] = team.NewTeam(scope, jsii.String(name), &team.TeamConfig{
			Name: jsii.String(name),
		})
	}

	return teams
}

func generateTeamMemberships(scope constructs.Construct, usersCSVData []UserRecord, users map[string]user.User, teams map[string]team.Team) map[string]teammembership.TeamMembership {
	teamMemberships := make(map[string]teammembership.TeamMembership)
	// template for "<team-name>_<user-name>"
	for _, u := range usersCSVData {
		tmName := fmt.Sprintf("%s_%s", u.Team, u.Key)
		teamMemberships[tmName] = teammembership.NewTeamMembership(scope, jsii.String(tmName), &teammembership.TeamMembershipConfig{
			TeamId: teams[u.Team].Id(),
			UserId: users[u.Email].Id(),
		})
	}

	return teamMemberships
}

func unique(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func main() {
	app := cdktf.NewApp(nil)

	NewMyStack(app, "terraform-provider-pagerduty-cdktf-local-testing")

	app.Synth()
}
