package iam

import (
	"fmt"

	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func CheckIf2FAActivated(checkConfig yatas.CheckConfig, mfaForUsers []MFAForUser, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("IAM Users have 2FA activated", "Check if all users have 2FA activated", testName)
	for _, mfaForUser := range mfaForUsers {
		if len(mfaForUser.MFAs) == 0 {
			Message := "2FA is not activated on " + mfaForUser.UserName
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: mfaForUser.UserName}
			check.AddResult(result)
		} else {
			Message := "2FA is activated on " + mfaForUser.UserName
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: mfaForUser.UserName}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
