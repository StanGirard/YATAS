package apigateway

import (
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/stangirard/yatas/internal/results"
	"github.com/stangirard/yatas/internal/yatas"
)

func TestCheckIfStagesProtectedByAcl(t *testing.T) {
	type args struct {
		checkConfig yatas.CheckConfig
		stages      []types.Stage
		testName    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test if stages are have protected by ACL",
			args: args{
				checkConfig: yatas.CheckConfig{
					Wg:    &sync.WaitGroup{},
					Queue: make(chan results.Check, 1),
				},
				stages: []types.Stage{
					{
						AccessLogSettings: &types.AccessLogSettings{
							DestinationArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:apigateway-access-logs:log-stream:test-api-stages-cloudwatch-logs"),
						},
						StageName: aws.String("test-stage"),
						WebAclArn: aws.String("arn:aws:execute-api:us-east-1:123456789012:test-api/test-stage/GET/test-path"),
					},
				},
				testName: "test-name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckIfStagesProtectedByAcl(tt.args.checkConfig, tt.args.stages, tt.args.testName)
			tt.args.checkConfig.Wg.Add(1)
			go func() {
				for check := range tt.args.checkConfig.Queue {
					if check.Id != tt.args.testName {
						t.Errorf("Check name is not equal to test name")
					}
					if check.Status != "OK" {
						t.Errorf("Check status is not equal to OK")
					}
					tt.args.checkConfig.Wg.Done()
				}
			}()
			tt.args.checkConfig.Wg.Wait()
		})
	}
}

func TestCheckIfStagesProtectedByAclFail(t *testing.T) {
	type args struct {
		checkConfig yatas.CheckConfig
		stages      []types.Stage
		testName    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test if stages are have protected by ACL",
			args: args{
				checkConfig: yatas.CheckConfig{
					Wg:    &sync.WaitGroup{},
					Queue: make(chan results.Check, 1),
				},
				stages: []types.Stage{
					{
						AccessLogSettings: &types.AccessLogSettings{
							DestinationArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:apigateway-access-logs:log-stream:test-api-stages-cloudwatch-logs"),
						},
						StageName: aws.String("test-stage"),
					},
				},
				testName: "test-name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckIfStagesProtectedByAcl(tt.args.checkConfig, tt.args.stages, tt.args.testName)
			tt.args.checkConfig.Wg.Add(1)
			go func() {
				for check := range tt.args.checkConfig.Queue {
					if check.Status != "FAIL" {
						t.Errorf("Check status is not equal to FAIL")
					}
					tt.args.checkConfig.Wg.Done()
				}
			}()
			tt.args.checkConfig.Wg.Wait()
		})
	}
}
