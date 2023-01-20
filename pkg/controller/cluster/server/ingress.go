package server

import (
	"context"

	"github.com/galal-hussein/k3k/pkg/apis/k3k.io/v1alpha1"
	"github.com/galal-hussein/k3k/pkg/controller/util"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	pathType    = networkingv1.PathTypePrefix
	wildcardDNS = ".sslip.io"
)

func Ingress(ctx context.Context, cluster *v1alpha1.Cluster, client client.Client) (*networkingv1.Ingress, error) {
	addresses, err := addresses(ctx, client)
	if err != nil {
		return nil, err
	}

	ingressRules := ingressRules(cluster, addresses)
	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-server-ingress",
			Namespace: util.ClusterNamespace(cluster),
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &cluster.Spec.IngressClassName,
			Rules:            ingressRules,
		},
	}, nil
}

// return all the nodes external addresses, if not found then return internal addresses
func addresses(ctx context.Context, client client.Client) ([]string, error) {
	addresses := []string{}
	nodeList := v1.NodeList{}
	if err := client.List(ctx, &nodeList); err != nil {
		return nil, err
	}

	for _, node := range nodeList.Items {
		addresses = append(addresses, GetNodeAddress(&node))
	}

	return addresses, nil
}

func GetNodeAddress(node *v1.Node) string {
	externalIP := ""
	internalIP := ""
	for _, ip := range node.Status.Addresses {
		if ip.Type == "ExternalIP" && ip.Address != "" {
			externalIP = ip.Address
			break
		} else if ip.Type == "InternalIP" && ip.Address != "" {
			internalIP = ip.Address
		}
	}
	if externalIP != "" {
		return externalIP
	}

	return internalIP
}

func ingressRules(cluster *v1alpha1.Cluster, addresses []string) []networkingv1.IngressRule {
	ingressRules := []networkingv1.IngressRule{}
	for _, address := range addresses {
		rule := networkingv1.IngressRule{
			Host: cluster.Name + "." + address + wildcardDNS,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: "k3k-server-service",
									Port: networkingv1.ServiceBackendPort{
										Number: 6443,
									},
								},
							},
						},
					},
				},
			},
		}
		ingressRules = append(ingressRules, rule)
	}
	return ingressRules
}