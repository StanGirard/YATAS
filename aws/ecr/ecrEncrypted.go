package ecr

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func CheckIfEncrypted(checkConfig yatas.CheckConfig, ecr []types.Repository, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("ECRs are encrypted", "Check if all ECRs are encrypted", testName)
	for _, ecr := range ecr {
		if ecr.EncryptionConfiguration == nil {
			Message := "ECR " + *ecr.RepositoryName + " is not encrypted"
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: *ecr.RepositoryName}
			check.AddResult(result)
		} else {
			Message := "ECR " + *ecr.RepositoryName + " is encrypted"
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: *ecr.RepositoryName}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
