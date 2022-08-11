package autoscaling

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/results"
	"github.com/stangirard/yatas/internal/yatas"
)

func GetAutoscalingGroups(s aws.Config) []types.AutoScalingGroup {
	svc := autoscaling.NewFromConfig(s)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	result, err := svc.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return nil
	}
	return result.AutoScalingGroups
}

func CheckIfDesiredCapacityMaxCapacityBelow80percent(wg *sync.WaitGroup, s aws.Config, groups []types.AutoScalingGroup, testName string, queueToAdd chan results.Check) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check results.Check
	check.Name = "Autoscaling DesiredCapacity MaxCapacity below 80%"
	check.Id = testName
	check.Description = "Check if all autoscaling groups have a desired capacity below 80%"
	check.Status = "OK"
	for _, group := range groups {
		if group.DesiredCapacity != nil && group.MaxSize != nil && float64(*group.DesiredCapacity) > float64(*group.MaxSize)*0.8 {
			check.Status = "FAIL"
			status := "FAIL"
			Message := "Autoscaling group " + *group.AutoScalingGroupName + " has a desired capacity above 80%"
			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *group.AutoScalingGroupName})
		} else {
			status := "OK"
			Message := "Autoscaling group " + *group.AutoScalingGroupName + " has a desired capacity below 80%"
			check.Results = append(check.Results, results.Result{Status: status, Message: Message, ResourceID: *group.AutoScalingGroupName})
		}
	}
	queueToAdd <- check
}

func RunChecks(wa *sync.WaitGroup, s aws.Config, c *yatas.Config, queue chan []results.Check) {

	var checks []results.Check
	groups := GetAutoscalingGroups(s)
	var wg sync.WaitGroup
	queueResults := make(chan results.Check, 10)
	go yatas.CheckTest(&wg, c, "AWS_ASG_001", CheckIfDesiredCapacityMaxCapacityBelow80percent)(&wg, s, groups, "AWS_ASG_001", queueResults)

	go func() {
		for t := range queueResults {
			checks = append(checks, t)
			wg.Done()
		}
	}()

	wg.Wait()

	queue <- checks
}
