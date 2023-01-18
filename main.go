//go:generate ./hack/update-codegen.sh
package main

import (
	"context"
	"flag"

	"github.com/galal-hussein/k3k/pkg/apis/k3k.io/v1alpha1"
	"github.com/galal-hussein/k3k/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	kubeconfig string
)

var (
	Scheme = runtime.NewScheme()
)

func init() {
	flag.StringVar(&kubeconfig, "kube-config", "", "kubeconfig path")
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = v1alpha1.AddToScheme(Scheme)
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to create config from kubeconfig file: %v", err)
	}

	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme: Scheme,
	})
	if err != nil {
		klog.Fatalf("Failed to create new controller runtime manager: %v", err)
	}
	if err := controller.Add(mgr); err != nil {
		klog.Fatalf("Failed to add the new controller: %v", err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		klog.Fatalf("Failed to start the manager: %v", err)
	}
}