/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	cmd "github.com/cita-cloud/cita-cloud-operator/pkg/exec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ChainNodeReconciler reconciles a ChainNode object
type ChainNodeReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	updateConfigFlag      bool
	updateStatefulSetFlag bool
	startUpdateTime       time.Time
	cmd                   cmd.Cmd
}

//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/finalizers,verbs=update

func (r *ChainNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("chainnode %s in reconcile", req.NamespacedName))

	chainNode := &citacloudv1.ChainNode{}
	if err := r.Get(ctx, req.NamespacedName, chainNode); err != nil {
		logger.Info(fmt.Sprintf("the chainnode %s has been deleted", req.NamespacedName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	chainConfig := &citacloudv1.ChainConfig{}
	if err := r.Get(ctx, types.NamespacedName{Name: chainNode.Spec.ChainName, Namespace: chainNode.Namespace}, chainConfig); err != nil {
		logger.Info(fmt.Sprintf("the chainconfig %s/%s has been deleted", req.Namespace, chainNode.Spec.ChainName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if chainConfig.Status.Status != citacloudv1.Online {
		// if chain is not online
		logger.Info(fmt.Sprintf("the chain [%s] is not online", chainConfig.Name))
		oldChainNode := chainNode.DeepCopy()
		chainNode.Status.Status = citacloudv1.NodeWaitChainOnline
		if !IsEqual(oldChainNode.Status, chainNode.Status) {
			if err := r.Client.Status().Update(ctx, chainNode); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info(fmt.Sprintf("update chain node status [%s] success", citacloudv1.NodeWaitChainOnline))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	if chainNode.Spec.Action == citacloudv1.NodeInitialize {
		oldChainNode := chainNode.DeepCopy()
		if err := r.SetDefaultSpec(ctx, chainConfig, chainNode); err != nil {
			return ctrl.Result{}, err
		}
		if !IsEqual(oldChainNode.Spec, chainNode.Spec) {
			diff, _ := DiffObject(oldChainNode, chainNode)
			logger.Info("SetDefault: " + string(diff))
			return ctrl.Result{}, r.Update(ctx, chainNode)
		}

		updated, err := r.SetDefaultStatus(ctx, chainNode)
		if updated || err != nil {
			return ctrl.Result{}, err
		}
	} else if chainNode.Spec.Action == citacloudv1.NodeStart {
		// create all resource
		if err := r.ReconcileAllRecourse(ctx, chainConfig, chainNode); err != nil {
			return ctrl.Result{}, err
		}
		// sync status
		if err := r.SyncRunningStatus(ctx, chainNode); err != nil {
			return ctrl.Result{}, err
		}
	} else if chainNode.Spec.Action == citacloudv1.NodeStop {
		_, err := r.ReconcileStatefulSet(ctx, chainConfig, chainNode)
		if err != nil {
			return ctrl.Result{}, err
		}
		// sync status
		if err := r.SyncStopStatus(ctx, chainNode); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ChainNodeReconciler) SetDefaultSpec(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)

	if chainNode.Spec.PullPolicy == "" {
		chainNode.Spec.PullPolicy = corev1.PullIfNotPresent
	}

	// set default images
	if chainNode.Spec.NetworkImage == "" {
		if chainConfig.Spec.EnableTLS {
			chainNode.Spec.NetworkImage = "citacloud/network_tls:v6.3.0"
		} else {
			chainNode.Spec.NetworkImage = "citacloud/network_p2p:v6.3.0"
		}
	}
	if chainNode.Spec.ConsensusImage == "" {
		if chainConfig.Spec.ConsensusType == citacloudv1.Raft {
			chainNode.Spec.ConsensusImage = "citacloud/consensus_raft:v6.3.0"
		} else if chainConfig.Spec.ConsensusType == citacloudv1.BFT {
			chainNode.Spec.ConsensusImage = "citacloud/consensus_bft:v6.3.0"
		} else {
			return fmt.Errorf("mismatched consensus type")
		}
	}
	if chainNode.Spec.ExecutorImage == "" {
		chainNode.Spec.ExecutorImage = "citacloud/executor_evm:v6.3.0"
	}
	if chainNode.Spec.StorageImage == "" {
		chainNode.Spec.StorageImage = "citacloud/storage_rocksdb:v6.3.0"
	}
	if chainNode.Spec.ControllerImage == "" {
		chainNode.Spec.ControllerImage = "citacloud/controller:v6.3.0"
	}
	if chainNode.Spec.KmsImage == "" {
		chainNode.Spec.KmsImage = "citacloud/kms_sm:v6.3.0"
	}

	// set domain
	if chainNode.Spec.Domain == "" {
		account := &citacloudv1.Account{}
		if err := r.Get(ctx, types.NamespacedName{Name: chainNode.Spec.Account, Namespace: chainNode.Namespace}, account); err != nil {
			logger.Error(err, fmt.Sprintf("get account [%s] failed", chainNode.Spec.Account))
			return err
		}
		chainNode.Spec.Domain = account.Spec.Domain
	}

	// set ownerReference
	if len(chainNode.OwnerReferences) == 0 {
		if err := ctrl.SetControllerReference(chainConfig, chainNode, r.Scheme); err != nil {
			logger.Error(err, "set chain node controller reference failed")
			return err
		}
		logger.Info("set chain node controller reference success")
	}
	return nil
}

func (r *ChainNodeReconciler) SetDefaultStatus(ctx context.Context, chainNode *citacloudv1.ChainNode) (bool, error) {
	logger := log.FromContext(ctx)
	if chainNode.Status.Status == "" {
		chainNode.Status.Status = citacloudv1.NodeInitialized
		err := r.Client.Status().Update(ctx, chainNode)
		if err != nil {
			return false, err
		}
		logger.Info("set chain node default status success")
		return true, nil
	}
	return false, nil
}

func (r *ChainNodeReconciler) ReconcileAllRecourse(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	var err error
	// reconcile service
	if err := r.ReconcileService(ctx, chainConfig, chainNode); err != nil {
		return err
	}
	// reconcile node config
	r.updateConfigFlag, err = r.ReconcileConfigMap(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	// reconcile log config
	r.updateConfigFlag, err = r.ReconcileLogConfigMap(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	// reconcile statefulset
	r.updateStatefulSetFlag, err = r.ReconcileStatefulSet(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger := log.FromContext(context.TODO())

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObj := e.ObjectOld.(*citacloudv1.ChainConfig)
			newObj := e.ObjectNew.(*citacloudv1.ChainConfig)
			if !reflect.DeepEqual(oldObj.Spec.ImageInfo, newObj.Spec.ImageInfo) {
				logger.Info(fmt.Sprintf("the chain [%s/%s] imageInfo field has changed, enqueue", newObj.Namespace, newObj.Name))
				return true
			}
			return false
		},
	}

	opts := []builder.WatchesOption{
		builder.WithPredicates(p),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.ChainNode{}).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(r.statefulSetPredicates())).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&source.Kind{Type: &citacloudv1.ChainConfig{}},
			//&handler.EnqueueRequestForOwner{OwnerType: &citacloudv1.ChainNode{}, IsController: false},
			handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
				// 筛选出下面所有的node节点，并进行入队
				reqs := make([]reconcile.Request, 0)
				chainConfig := a.(*citacloudv1.ChainConfig)
				for name := range chainConfig.Status.NodeInfoMap {
					req := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: a.GetNamespace()}}
					reqs = append(reqs, req)
				}
				return reqs
			}),
			opts...,
		).
		Complete(r)
}

func (r *ChainNodeReconciler) statefulSetPredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(event event.UpdateEvent) bool {
			curSet := event.ObjectNew.(*appsv1.StatefulSet)
			oldSet := event.ObjectOld.(*appsv1.StatefulSet)
			if reflect.DeepEqual(curSet.Status, oldSet.Status) {
				return false
			}
			return true
		},
	}
}
