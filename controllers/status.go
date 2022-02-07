package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SyncRunningStatus
// 如果status == Initialized，则判断当前pod的ready
func (r *ChainNodeReconciler) SyncRunningStatus(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	//var updateFlag bool
	oldStatus := chainNode.Status.DeepCopy()

	sts := &v1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, sts)
	if errors.IsNotFound(err) {
		chainNode.Status.Status = citacloudv1.NodeStarting
		if !reflect.DeepEqual(oldStatus, chainNode.Status) {
			logger.Info("updating chain node status [Starting]...")
			return r.Status().Update(ctx, chainNode)
		}
		return nil
	} else if err != nil {
		return err
	}

	if chainNode.Status.Status == citacloudv1.NodeInitialized {
		chainNode.Status.Status = citacloudv1.NodeStarting
	} else if chainNode.Status.Status == citacloudv1.NodeStarting {
		if pointer.Int32Equal(pointer.Int32(sts.Status.ReadyReplicas), sts.Spec.Replicas) {
			chainNode.Status.Status = citacloudv1.NodeRunning
		}
	} else if chainNode.Status.Status == citacloudv1.NodeRunning {
		if !pointer.Int32Equal(pointer.Int32(sts.Status.ReadyReplicas), sts.Spec.Replicas) {
			chainNode.Status.Status = citacloudv1.NodeError
		} else if !reflect.DeepEqual(chainConfig.Spec.ImageInfo, chainNode.Spec.ImageInfo) {
			// imageInfo is inconsistent
			chainNode.Status.Status = citacloudv1.NodeWarning
		}
	} else if chainNode.Status.Status == citacloudv1.NodeError {
		if pointer.Int32Equal(pointer.Int32(sts.Status.ReadyReplicas), sts.Spec.Replicas) {
			chainNode.Status.Status = citacloudv1.NodeRunning
		}
	} else if chainNode.Status.Status == citacloudv1.NodeUpdating {
		// Check again 5 seconds after the update starts, what a stupid way
		if time.Since(r.startUpdateTime).Seconds() > 5 {
			if pointer.Int32Equal(pointer.Int32(sts.Status.ReadyReplicas), sts.Spec.Replicas) && sts.Status.CurrentRevision == sts.Status.UpdateRevision {
				chainNode.Status.Status = citacloudv1.NodeRunning
				r.updateStatefulSetFlag = false
			}
		}
	} else if chainNode.Status.Status == citacloudv1.NodeStopped {
		chainNode.Status.Status = citacloudv1.NodeStarting
		r.updateStatefulSetFlag = false
	}

	if r.updateStatefulSetFlag {
		// statefulset has modified, set updating status
		r.startUpdateTime = time.Now()
		chainNode.Status.Status = citacloudv1.NodeUpdating
	}

	if r.updateConfigFlag {
		// chainnode's config modified, set warning status
		chainNode.Status.Status = citacloudv1.NodeWarning
	}

	currentStatus := chainNode.Status.DeepCopy()
	if !reflect.DeepEqual(oldStatus, currentStatus) {
		logger.Info(fmt.Sprintf("updating chain node status from [%s] to [%s]...", oldStatus.Status, currentStatus.Status))
		err = r.Status().Update(ctx, chainNode)
		if err != nil {
			logger.Error(err, fmt.Sprintf("update chain node status [%s] failed", currentStatus.Status))
			return err
		}
		logger.Info(fmt.Sprintf("update chain node status [%s] success", currentStatus.Status))
		return nil
	}
	logger.Info("chain node status has not changed")
	return nil
}

func (r *ChainNodeReconciler) SyncStopStatus(ctx context.Context, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	//var updateFlag bool
	oldStatus := chainNode.Status.DeepCopy()

	sts := &v1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, sts)
	if err != nil {
		return err
	}
	if chainNode.Status.Status != citacloudv1.NodeStopped {
		chainNode.Status.Status = citacloudv1.NodeStopping
	}
	if chainNode.Status.Status == citacloudv1.NodeStopping {
		if sts.Status.CurrentReplicas == 0 && sts.Status.Replicas == 0 {
			chainNode.Status.Status = citacloudv1.NodeStopped
		}
	}

	currentStatus := chainNode.Status.DeepCopy()
	if !reflect.DeepEqual(oldStatus, currentStatus) {
		logger.Info(fmt.Sprintf("updating chain node status from [%s] to [%s]...", oldStatus.Status, currentStatus.Status))
		err = r.Status().Update(ctx, chainNode)
		if err != nil {
			logger.Error(err, fmt.Sprintf("update chain node status [%s] failed", currentStatus.Status))
			return err
		}
		logger.Info(fmt.Sprintf("update chain node status [%s] success", currentStatus.Status))
		return nil
	}
	logger.Info("chain node status has not changed")
	return nil
}
