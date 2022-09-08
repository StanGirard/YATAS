package lambda

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func CheckIfLambdaPrivate(checkConfig yatas.CheckConfig, lambdas []types.FunctionConfiguration, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("Lambdas are private", "Check if all Lambdas are private", testName)
	for _, lambda := range lambdas {
		if lambda.VpcConfig == nil {
			Message := "Lambda " + *lambda.FunctionName + " is public"
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: *lambda.FunctionArn}
			check.AddResult(result)
		} else {
			Message := "Lambda " + *lambda.FunctionName + " is private"
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: *lambda.FunctionArn}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
