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

	generateUsers(stack, usersData)
	generateTeams(stack, usersData)

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

func generateUsers(scope constructs.Construct, usersData []UserRecord) []user.User {
	users := []user.User{}
	for _, u := range usersData {
		newUser := user.NewUser(scope, jsii.String(u.Key), &user.UserConfig{
			Name:     jsii.String(u.Name),
			Email:    jsii.String(u.Email),
			Role:     jsii.String(u.Role),
			JobTitle: jsii.String(u.JobTitle),
		})
		users = append(users, newUser)
	}

	return users
}

func generateTeams(scope constructs.Construct, usersData []UserRecord) []team.Team {
	teams := []team.Team{}

	return teams
}

func main() {
	app := cdktf.NewApp(nil)

	NewMyStack(app, "terraform-provider-pagerduty-cdktf-local-testing")

	app.Synth()
}
