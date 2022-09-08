package volumes

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stangirard/yatas/internal/logger"
	"github.com/stangirard/yatas/internal/yatas"
)

func CheckIfAllSnapshotsEncrypted(checkConfig yatas.CheckConfig, snapshots []types.Snapshot, testName string) {
	logger.Info(fmt.Sprint("Running ", testName))
	var check yatas.Check
	check.InitCheck("EC2's Snapshots are encrypted", "Check if all snapshots are encrypted", testName)
	for _, snapshot := range snapshots {
		if snapshot.Encrypted == nil || !*snapshot.Encrypted {
			Message := "Snapshot " + *snapshot.SnapshotId + " is not encrypted"
			result := yatas.Result{Status: "FAIL", Message: Message, ResourceID: *snapshot.SnapshotId}
			check.AddResult(result)
		} else {
			Message := "Snapshot " + *snapshot.SnapshotId + " is encrypted"
			result := yatas.Result{Status: "OK", Message: Message, ResourceID: *snapshot.SnapshotId}
			check.AddResult(result)
		}
	}
	checkConfig.Queue <- check
}
