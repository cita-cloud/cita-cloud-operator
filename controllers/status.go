package controllers

import (
	"context"
	"fmt"
	"reflect"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SyncStatus
// 如果status == Initialized，则判断当前pod的ready
func (r *ChainNodeReconciler) SyncStatus(ctx context.Context, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	//var updateFlag bool
	oldStatus := chainNode.Status.DeepCopy()

	sts := &v1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, sts)
	if errors.IsNotFound(err) {
		chainNode.Status.Status = citacloudv1.NodeCreating
		if !reflect.DeepEqual(oldStatus, chainNode.Status) {
			logger.Info("updating chain node status [Creating]...")
			return r.Status().Update(ctx, chainNode)
		}
		return nil
	} else if err != nil {
		return err
	}

	if chainNode.Status.Status == citacloudv1.NodeInitialized {
		chainNode.Status.Status = citacloudv1.NodeCreating
	} else if chainNode.Status.Status == citacloudv1.NodeCreating {
		if sts.Status.ReadyReplicas == sts.Status.Replicas {
			chainNode.Status.Status = citacloudv1.NodeRunning
		}
	} else if chainNode.Status.Status == citacloudv1.NodeRunning {
		if sts.Status.ReadyReplicas != sts.Status.Replicas {
			chainNode.Status.Status = citacloudv1.NodeError
		}
	} else if chainNode.Status.Status == citacloudv1.NodeError {
		if sts.Status.ReadyReplicas == sts.Status.Replicas {
			chainNode.Status.Status = citacloudv1.NodeRunning
		}
	} else if chainNode.Status.Status == citacloudv1.NodeUpdating {
		// todo: set updating status
		if sts.Status.ReadyReplicas == sts.Status.Replicas {
			chainNode.Status.Status = citacloudv1.NodeRunning
		}
	}

	currentStatus := chainNode.Status.DeepCopy()
	if !reflect.DeepEqual(oldStatus, currentStatus) {
		logger.Info(fmt.Sprintf("updating chain node status from [%s] to [%s]...", oldStatus.Status, currentStatus.Status))
		return r.Status().Update(ctx, chainNode)
	}
	logger.Info("chain node status has not changed")
	return nil
}
