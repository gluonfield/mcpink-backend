package k8sdeployments

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/customdomains"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (a *Activities) ApplyCustomDomainIngress(ctx context.Context, input ApplyCustomDomainIngressInput) error {
	a.logger.Info("ApplyCustomDomainIngress",
		"namespace", input.Namespace,
		"serviceName", input.ServiceName,
		"domain", input.Domain)

	ingress := buildCustomDomainIngress(input.Namespace, input.ServiceName, input.Domain, input.Port)
	data, err := json.Marshal(ingress)
	if err != nil {
		return fmt.Errorf("marshal custom domain ingress: %w", err)
	}

	ingressName := input.ServiceName + "-cd"
	_, err = a.k8s.NetworkingV1().Ingresses(input.Namespace).Patch(ctx, ingressName,
		types.ApplyPatchType, data,
		metav1.PatchOptions{FieldManager: "temporal-worker"})
	if err != nil {
		return fmt.Errorf("apply custom domain ingress: %w", err)
	}

	return nil
}

func (a *Activities) DeleteCustomDomainIngress(ctx context.Context, input DeleteCustomDomainIngressInput) error {
	a.logger.Info("DeleteCustomDomainIngress",
		"namespace", input.Namespace,
		"serviceName", input.ServiceName)

	ingressName := input.ServiceName + "-cd"
	tlsSecretName := ingressName + "-tls"

	err := a.k8s.NetworkingV1().Ingresses(input.Namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete custom domain ingress: %w", err)
	}

	err = a.k8s.CoreV1().Secrets(input.Namespace).Delete(ctx, tlsSecretName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete custom domain tls secret: %w", err)
	}

	return nil
}

func (a *Activities) UpdateCustomDomainDBStatus(ctx context.Context, input UpdateCustomDomainStatusInput) error {
	a.logger.Info("UpdateCustomDomainDBStatus",
		"customDomainID", input.CustomDomainID,
		"status", input.Status)

	if input.Status == "active" {
		_, err := a.customDomainsQ.UpdateVerified(ctx, input.CustomDomainID)
		return err
	}

	_, err := a.customDomainsQ.UpdateStatus(ctx, customdomains.UpdateStatusParams{
		ID:     input.CustomDomainID,
		Status: input.Status,
	})
	return err
}
