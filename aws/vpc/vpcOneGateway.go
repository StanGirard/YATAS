package vpc

import (
	"fmt"

	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func checkIfOnlyOneGateway(checkConfig yatas.CheckConfig, vpcInternetGateways []VpcToInternetGateway, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("VPC only have one Gateway", "Check if VPC has only one gateway", testName)
	for _, vpcInternetGateway := range vpcInternetGateways {
		if len(vpcInternetGateway.InternetGateways) > 1 {
			Message := "VPC has more than one gateway on " + vpcInternetGateway.VpcID
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: vpcInternetGateway.VpcID}
			check.AddResult(result)
		} else {
			Message := "VPC has only one gateway on " + vpcInternetGateway.VpcID
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: vpcInternetGateway.VpcID}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
