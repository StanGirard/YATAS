package iam

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/results"
	"github.com/stangirard/yatas/internal/yatas"
)

func GetAllUsers(s aws.Config) []types.User {
	svc := iam.NewFromConfig(s)
	input := &iam.ListUsersInput{}
	result, err := svc.ListUsers(context.TODO(), input)
	if err != nil {
		panic(err)
	}
	return result.Users
}

func CheckIf2FAActivated(wg *sync.WaitGroup, s aws.Config, users []types.User, testName string, queueToAdd chan results.Check) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check results.Check
	check.Name = "IAM 2FA"
	check.Id = testName
	check.Description = "Check if all users have 2FA activated"
	check.Status = "OK"
	svc := iam.NewFromConfig(s)
	for _, user := range users {
		// List MFA devices for the user
		params := &iam.ListMFADevicesInput{
			UserName: user.UserName,
		}
		resp, err := svc.ListMFADevices(context.TODO(), params)
		if err != nil {
			panic(err)
		}
		if len(resp.MFADevices) == 0 {
			check.Status = "FAIL"
			status := "FAIL"
			Message := "2FA is not activated on " + *user.UserName
			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *user.UserName})
		} else {
			status := "OK"
			Message := "2FA is activated on " + *user.UserName
			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *user.UserName})
		}
	}
	queueToAdd <- check
}

func CheckAgeAccessKeyLessThan90Days(wg *sync.WaitGroup, s aws.Config, users []types.User, testName string, queueToAdd chan results.Check) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check results.Check
	check.Name = "IAM Access Key Age"
	check.Id = testName
	check.Description = "Check if all users have access key less than 90 days"
	check.Status = "OK"
	svc := iam.NewFromConfig(s)
	for _, user := range users {
		// List access keys for the user
		params := &iam.ListAccessKeysInput{
			UserName: user.UserName,
		}
		resp, err := svc.ListAccessKeys(context.TODO(), params)
		if err != nil {
			panic(err)
		}
		now := time.Now()
		for _, accessKey := range resp.AccessKeyMetadata {
			if now.Sub(*accessKey.CreateDate).Hours() > 2160 {
				check.Status = "FAIL"
				status := "FAIL"
				Message := "Access key " + *accessKey.AccessKeyId + " is older than 90 days on " + *user.UserName
				check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *user.UserName})
			} else {
				status := "OK"
				Message := "Access key " + *accessKey.AccessKeyId + " is younger than 90 days on " + *user.UserName
				check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *user.UserName})
			}
		}
	}
	queueToAdd <- check
}

type UserPolicies struct {
	UserName string
	Policies []Policy
}

func CheckIfUserCanElevateRights(wg *sync.WaitGroup, s aws.Config, users []types.User, testName string, queueToAdd chan results.Check) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check results.Check
	check.Name = "IAM User Can Elevate Rights"
	check.Id = testName
	check.Description = "Check if  users can elevate rights"
	check.Status = "OK"
	var wgPolicyForUser sync.WaitGroup
	queue := make(chan UserPolicies, len(users))

	for _, user := range users {
		go GetAllPolicyForUser(&wgPolicyForUser, queue, s, user)
	}
	var userPolicies []UserPolicies
	go func() {
		for user := range queue {
			userPolicies = append(userPolicies, user)
			wgPolicyForUser.Done()
		}

	}()
	wgPolicyForUser.Wait()
	for _, userPol := range userPolicies {
		elevation := CheckPolicyForAllowInRequiredPermission(userPol.Policies, requiredPermissions)
		if len(elevation) > 0 {
			check.Status = "FAIL"
			status := "FAIL"
			var Message string
			if len(elevation) > 3 {
				Message = "User " + userPol.UserName + " can elevate rights with " + fmt.Sprint(elevation[len(elevation)-3:]) + " only last 3 policies"
			} else {
				Message = "User " + userPol.UserName + " can elevate rights with " + fmt.Sprint(elevation)
			}

			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: userPol.UserName})
		} else {
			status := "OK"
			Message := "User " + userPol.UserName + " cannot elevate rights"
			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: userPol.UserName})
		}
	}
	queueToAdd <- check
}

func RunChecks(wa *sync.WaitGroup, s aws.Config, c *yatas.Config, queue chan []results.Check) {

	var checks []results.Check
	users := GetAllUsers(s)
	var wg sync.WaitGroup
	queueResults := make(chan results.Check, 10)

	go yatas.CheckTest(&wg, c, "AWS_IAM_001", CheckIf2FAActivated)(&wg, s, users, "AWS_IAM_001", queueResults)
	go yatas.CheckTest(&wg, c, "AWS_IAM_002", CheckAgeAccessKeyLessThan90Days)(&wg, s, users, "AWS_IAM_002", queueResults)
	go yatas.CheckTest(&wg, c, "AWS_IAM_003", CheckIfUserCanElevateRights)(&wg, s, users, "AWS_IAM_003", queueResults)
	go func() {
		for t := range queueResults {
			checks = append(checks, t)
			wg.Done()
		}
	}()

	wg.Wait()

	queue <- checks
}
