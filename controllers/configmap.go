package controllers

import (
	"context"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ChainNodeReconciler) ReconcileConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	old := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: GetNodeConfigName(chainNode.Spec.ChainName, chainNode.Name), Namespace: chainNode.Namespace}, old)
	if errors.IsNotFound(err) {
		newObj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetNodeConfigName(chainNode.Spec.ChainName, chainNode.Name),
				Namespace: chainNode.Namespace,
			},
		}
		if err = r.updateNodeConfigMap(ctx, chainConfig, chainNode, newObj); err != nil {
			return err
		}
		logger.Info("create node config configmap....")
		return r.Create(ctx, newObj)
	} else if err != nil {
		return err
	}

	cur := old.DeepCopy()
	if err := r.updateNodeConfigMap(ctx, chainConfig, chainNode, cur); err != nil {
		return err
	}
	if IsEqual(old, cur) {
		logger.Info("the configmap part has not changed, go pass")
		return nil
	}

	logger.Info("update node configmap...")
	return r.Update(ctx, cur)
}

func (r *ChainNodeReconciler) updateNodeConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, configMap *corev1.ConfigMap) error {
	logger := log.FromContext(ctx)
	configMap.Labels = MergeLabels(configMap.Labels, LabelsForNode(chainNode.Spec.ChainName, chainNode.Name))
	if err := ctrl.SetControllerReference(chainNode, configMap, r.Scheme); err != nil {
		logger.Error(err, "node configmap SetControllerReference error")
		return err
	}

	cnService := NewChainNodeService(chainConfig, chainNode)

	configMap.Data = map[string]string{
		NodeConfigFile: cnService.GenerateNodeConfig(),
	}
	return nil
}

func (r *ChainNodeReconciler) ReconcileLogConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	old := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: GetLogConfigName(chainNode.Spec.ChainName, chainNode.Name), Namespace: chainNode.Namespace}, old)
	if errors.IsNotFound(err) {
		newObj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetLogConfigName(chainNode.Spec.ChainName, chainNode.Name),
				Namespace: chainNode.Namespace,
			},
		}
		if err = r.updateLogConfigMap(ctx, chainConfig, chainNode, newObj); err != nil {
			return err
		}
		logger.Info("create log config configmap....")
		return r.Create(ctx, newObj)
	} else if err != nil {
		return err
	}

	cur := old.DeepCopy()
	if err := r.updateLogConfigMap(ctx, chainConfig, chainNode, cur); err != nil {
		return err
	}
	if IsEqual(old, cur) {
		logger.Info("the log configmap part has not changed, go pass")
		return nil
	}

	logger.Info("update log configmap...")
	return r.Update(ctx, cur)
}

func (r *ChainNodeReconciler) updateLogConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, configMap *corev1.ConfigMap) error {
	logger := log.FromContext(ctx)
	configMap.Labels = MergeLabels(configMap.Labels, LabelsForNode(chainNode.Spec.ChainName, chainNode.Name))
	if err := ctrl.SetControllerReference(chainNode, configMap, r.Scheme); err != nil {
		logger.Error(err, "log configmap SetControllerReference error")
		return err
	}

	cnService := NewChainNodeService(chainConfig, chainNode)

	configMap.Data = map[string]string{
		ControllerLogConfigFile: cnService.GenerateControllerLogConfig(),
		ExecutorLogConfigFile:   cnService.GenerateExecutorLogConfig(),
		KmsLogConfigFile:        cnService.GenerateKmsLogConfig(),
		NetworkLogConfigFile:    cnService.GenerateNetworkLogConfig(),
		StorageLogConfigFile:    cnService.GenerateStorageLogConfig(),
	}
	return nil
}
