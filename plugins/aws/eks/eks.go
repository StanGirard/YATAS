package eks

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stangirard/yatas/internal/yatas"
)

func RunChecks(wa *sync.WaitGroup, s aws.Config, c *yatas.Config, queue chan []yatas.Check) {
	var checkConfig yatas.CheckConfig
	checkConfig.Init(s, c)
	var checks []yatas.Check
	svc := eks.NewFromConfig(s)
	clusters := GetClusters(svc)
	clusterToUpdate := GetUpdates(svc, clusters)
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_EKS_001", CheckIfLoggingIsEnabled)(checkConfig, clusters, "AWS_EKS_001")
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_EKS_002", CheckIfEksEndpointPrivate)(checkConfig, clusters, "AWS_EKS_002")
	go yatas.CheckTest(checkConfig.Wg, c, "AWS_EKS_003", CheckIfEKSUpdateAvailable)(checkConfig, clusterToUpdate, "AWS_EKS_003")
	go func() {
		for t := range checkConfig.Queue {
			t.EndCheck()
			checks = append(checks, t)
			if c.CheckProgress.Bar != nil {
				c.CheckProgress.Bar.Increment()
			}

			checkConfig.Wg.Done()

		}
	}()
	checkConfig.Wg.Wait()
	queue <- checks
}
