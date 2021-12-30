package controllers

import (
	"context"
	"fmt"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ChainNodeReconciler) ReconcileConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (bool, error) {
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
			return false, err
		}
		logger.Info("create node config configmap....")
		return false, r.Create(ctx, newObj)
	} else if err != nil {
		return false, err
	}

	cur := old.DeepCopy()
	if err := r.updateNodeConfigMap(ctx, chainConfig, chainNode, cur); err != nil {
		return false, err
	}
	if IsEqual(old, cur) {
		logger.Info("the configmap part has not changed, go pass")
		return false, nil
	}

	logger.Info("update node configmap...")
	return true, r.Update(ctx, cur)
}

func (r *ChainNodeReconciler) updateNodeConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, configMap *corev1.ConfigMap) error {
	logger := log.FromContext(ctx)
	configMap.Labels = MergeLabels(configMap.Labels, LabelsForNode(chainNode.Spec.ChainName, chainNode.Name))
	if err := ctrl.SetControllerReference(chainNode, configMap, r.Scheme); err != nil {
		logger.Error(err, "node configmap SetControllerReference error")
		return err
	}
	var cnService *ChainNodeService

	// find account
	account:= &citacloudv1.Account{}
	if err := r.Get(ctx, types.NamespacedName{Name: chainNode.Spec.Account, Namespace: chainNode.Namespace}, account); err != nil {
		logger.Error(err, fmt.Sprintf("get account [%s] failed", chainNode.Spec.Account))
		return err
	}

	if chainConfig.Spec.EnableTLS {
		// todo: reflect
		// get chain ca secret
		caSecret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: GetCaSecretName(chainConfig.Name), Namespace: chainConfig.Namespace}, caSecret); err != nil {
			logger.Error(err, "get chain secret error")
			return err
		}
		// get node secret
		nodeCertAndKeySecret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: GetAccountCertAndKeySecretName(chainConfig.Name, chainNode.Spec.Account), Namespace: chainNode.Namespace}, nodeCertAndKeySecret); err != nil {
			logger.Error(err, "get node secret error")
			return err
		}

		cnService = NewChainNodeServiceForTls(chainConfig, chainNode, account, caSecret, nodeCertAndKeySecret)
	} else {
		cnService = NewChainNodeServiceForP2P(chainConfig, chainNode, account)
	}

	configMap.Data = map[string]string{
		NodeConfigFile: cnService.GenerateNodeConfig(),
	}
	return nil
}

func (r *ChainNodeReconciler) ReconcileLogConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (bool, error) {
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
			return false, err
		}
		logger.Info("create log config configmap....")
		return false, r.Create(ctx, newObj)
	} else if err != nil {
		return false, err
	}

	cur := old.DeepCopy()
	if err := r.updateLogConfigMap(ctx, chainConfig, chainNode, cur); err != nil {
		return false, err
	}
	if IsEqual(old, cur) {
		logger.Info("the log configmap part has not changed, go pass")
		return false, nil
	}

	logger.Info("update log configmap...")
	return true, r.Update(ctx, cur)
}

func (r *ChainNodeReconciler) updateLogConfigMap(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, configMap *corev1.ConfigMap) error {
	logger := log.FromContext(ctx)
	configMap.Labels = MergeLabels(configMap.Labels, LabelsForNode(chainNode.Spec.ChainName, chainNode.Name))
	if err := ctrl.SetControllerReference(chainNode, configMap, r.Scheme); err != nil {
		logger.Error(err, "log configmap SetControllerReference error")
		return err
	}

	cnService := NewChainNodeServiceForLog(chainConfig, chainNode)

	configMap.Data = map[string]string{
		ControllerLogConfigFile: cnService.GenerateControllerLogConfig(),
		ExecutorLogConfigFile:   cnService.GenerateExecutorLogConfig(),
		KmsLogConfigFile:        cnService.GenerateKmsLogConfig(),
		NetworkLogConfigFile:    cnService.GenerateNetworkLogConfig(),
		StorageLogConfigFile:    cnService.GenerateStorageLogConfig(),
	}
	return nil
}
