package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stangirard/yatas/internal/aws/apigateway"
	"github.com/stangirard/yatas/internal/aws/autoscaling"
	"github.com/stangirard/yatas/internal/aws/cloudfront"
	"github.com/stangirard/yatas/internal/aws/cloudtrail"
	"github.com/stangirard/yatas/internal/aws/dynamodb"
	"github.com/stangirard/yatas/internal/aws/ec2"
	"github.com/stangirard/yatas/internal/aws/ecr"
	"github.com/stangirard/yatas/internal/aws/iam"
	"github.com/stangirard/yatas/internal/aws/lambda"
	"github.com/stangirard/yatas/internal/aws/loadbalancers"
	"github.com/stangirard/yatas/internal/aws/rds"
	"github.com/stangirard/yatas/internal/aws/s3"
	"github.com/stangirard/yatas/internal/aws/volumes"
	"github.com/stangirard/yatas/internal/aws/vpc"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/results"
	"github.com/stangirard/yatas/internal/yatas"
)

func Run(c *yatas.Config) ([]results.Check, error) {
	s := initAuth(c)
	logger.Info("Launching AWS checks")
	checks := initTest(s, c)
	return checks, nil
}

func initTest(s aws.Config, c *yatas.Config) []results.Check {

	var checks []results.Check
	checks = append(checks, s3.RunS3Test(s, c)...)
	checks = append(checks, volumes.RunVolumesTest(s, c)...)
	checks = append(checks, rds.RunRDSTests(s, c)...)
	checks = append(checks, vpc.RunVPCTests(s, c)...)
	checks = append(checks, cloudtrail.RunCloudtrailTests(s, c)...)
	checks = append(checks, ecr.RunECRTests(s, c)...)
	checks = append(checks, lambda.RunLambdaTests(s, c)...)
	checks = append(checks, dynamodb.RunDynamodbTests(s, c)...)
	checks = append(checks, ec2.RunEC2Tests(s, c)...)
	checks = append(checks, iam.RunIAMTests(s, c)...)
	checks = append(checks, cloudfront.RunCloudFrontTests(s, c)...)
	checks = append(checks, apigateway.RunApiGatewayTests(s, c)...)
	checks = append(checks, autoscaling.RunAutoscalingGroupChecks(s, c)...)
	checks = append(checks, loadbalancers.RunLoadBalancersTests(s, c)...)
	logger.Info("AWS checks completed ✅")

	return checks
}
