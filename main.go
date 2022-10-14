package main

import (
	"flag"
	"os"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/controller"
	"ctx.sh/apex-operator/pkg/scraper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	setupLog = ctrl.Log.WithName("setup")
	scheme   = runtime.NewScheme()
)

func init() {
	_ = apexv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

}

func main() {
	flag.Parse()

	ctx := ctrl.SetupSignalHandler()

	ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconciler := &controller.Reconciler{
		Client:   mgr.GetClient(),
		Log:      mgr.GetLogger().WithValues("controller", "apex"),
		Scrapers: scraper.NewManager(),
	}

	err = reconciler.SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	if os.Getenv("APEX_ENABLE_WEBHOOKS") != "false" {
		err = (&apexv1.Scraper{}).SetupWebhookWithManager(mgr)
		if err != nil {
			setupLog.Error(err, "Unable to setup webhooks", "webhook", "Scraper")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
