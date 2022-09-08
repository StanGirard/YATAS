package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func checkIfAutoUpgradeEnabled(checkConfig yatas.CheckConfig, instances []types.DBInstance, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("RDS have minor versions automatically updated", "Check if RDS minor auto upgrade is enabled", testName)
	for _, instance := range instances {
		if !instance.AutoMinorVersionUpgrade {
			Message := "RDS auto upgrade is not enabled on " + *instance.DBInstanceIdentifier
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: *instance.DBInstanceArn}
			check.AddResult(result)
		} else {
			Message := "RDS auto upgrade is enabled on " + *instance.DBInstanceIdentifier
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: *instance.DBInstanceArn}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
