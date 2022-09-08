package acm

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func CheckIfCertificateExpiresIn90Days(checkConfig yatas.CheckConfig, certificates []types.CertificateDetail, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("ACM certificate expires in more than 90 days", "Check if certificate expires in 90 days", testName)
	for _, certificate := range certificates {
		if certificate.Status == types.CertificateStatusIssued || certificate.Status == types.CertificateStatusInactive {
			if time.Until(*certificate.NotAfter).Hours() > 24*90 {
				Message := "Certificate " + *certificate.CertificateArn + " does not expire in 90 days"
				result := yatas.Result{Status: "OK", Message: Message, ResourceID: *certificate.CertificateArn}
				check.AddResult(result)
			} else {
				Message := "Certificate " + *certificate.CertificateArn + " expires in 90 days or less"
				result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: *certificate.CertificateArn}
				check.AddResult(result)
			}
		}
	}
	checkConfig.Queue <- check
}
