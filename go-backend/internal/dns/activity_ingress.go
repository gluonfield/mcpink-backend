package dns

import (
	"context"
	"encoding/json"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	networkingv1 "k8s.io/api/networking/v1"
)

func (a *Activities) ApplySubdomainIngress(ctx context.Context, input ApplySubdomainIngressInput) error {
	a.logger.Info("ApplySubdomainIngress",
		"namespace", input.Namespace,
		"serviceName", input.ServiceName,
		"fqdn", input.FQDN)

	ingress := buildSubdomainIngress(
		input.Namespace,
		input.ServiceName,
		input.FQDN,
		input.CertSecret,
		input.ServicePort,
	)

	data, err := json.Marshal(ingress)
	if err != nil {
		return fmt.Errorf("marshal subdomain ingress: %w", err)
	}

	ingressName := input.ServiceName + "-dz"
	_, err = a.k8s.NetworkingV1().Ingresses(input.Namespace).Patch(ctx, ingressName,
		types.ApplyPatchType, data,
		metav1.PatchOptions{FieldManager: "temporal-worker"})
	if err != nil {
		return fmt.Errorf("apply subdomain ingress: %w", err)
	}

	return nil
}

func (a *Activities) DeleteIngress(ctx context.Context, input DeleteIngressInput) error {
	a.logger.Info("DeleteIngress", "namespace", input.Namespace, "ingressName", input.IngressName)

	err := a.k8s.NetworkingV1().Ingresses(input.Namespace).Delete(ctx, input.IngressName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete ingress: %w", err)
	}
	return nil
}

func buildSubdomainIngress(namespace, serviceName, fqdn, certSecretName string, port int32) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingressClassName := "traefik"
	ingressName := serviceName + "-dz"

	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{Kind: "Ingress", APIVersion: "networking.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{fqdn},
					SecretName: certSecretName,
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: fqdn,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
