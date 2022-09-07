package apigateway

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/stangirard/yatas/internal/yatas"
)

func RunChecks(wa *sync.WaitGroup, s aws.Config, c *yatas.Config, queue chan []yatas.Check) {
	var checkConfig yatas.CheckConfig
	checkConfig.Init(s, c)
	var checks []yatas.Check
	svc := apigateway.NewFromConfig(s)
	apis := GetApiGateways(svc)
	stages := GetAllStagesApiGateway(svc, apis)
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_APG_001", CheckIfStagesCloudwatchLogsExist)(checkConfig, stages, "AWS_APG_001")
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_APG_002", CheckIfStagesProtectedByAcl)(checkConfig, stages, "AWS_APG_002")
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_APG_003", CheckIfTracingEnabled)(checkConfig, stages, "AWS_APG_003")

	go func() {
		for t := range checkConfig.Queue {
			t.EndCheck()
			checks = append(checks, t)

			checkConfig.Wg.Done()

		}
	}()

	checkConfig.Wg.Wait()

	queue <- checks
}
