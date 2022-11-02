package main

import (
	"flag"
	"os"

	"ctx.sh/apex"
	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/controller"
	"ctx.sh/apex-operator/pkg/scraper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DefaultCertDir              string = "/etc/admission-webhook/tls"
	DefaultEnableLeaderElection bool   = false
)

var (
	setupLog       = ctrl.Log.WithName("setup")
	scheme         = runtime.NewScheme()
	certDir        string
	leaderElection bool
)

func init() {
	_ = apexv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	flag.StringVar(&certDir, "certs", DefaultCertDir, "specify the cert directory")
	flag.BoolVar(&leaderElection, "enable-leader-election", DefaultEnableLeaderElection, "enable leader election")
}

func main() {
	flag.Parse()

	ctx := ctrl.SetupSignalHandler()

	metrics := apex.New(apex.MetricsOpts{
		Separator: '_',
		// make me configurable
		Port:           9090,
		ConstantLabels: []string{"controller", "apex.ctx.sh"},
		PanicOnError:   true,
	})
	// Need handle starts for apex-go a bit better, allow context
	// and orchestrate shutdown better.
	go func() {
		err := metrics.Start()
		if err != nil {
			setupLog.Error(err, "unable to start prometheus")
			os.Exit(1)
		}
	}()

	ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Port:   9443,
		// I was hoping that controller-runtime would give access
		// to the underlying registry or registerer to get any potential
		// builtin metrics (don't know if there is any tbh), but there
		// is no access, so take it over.
		MetricsBindAddress: "0",
		LeaderElection:     leaderElection,
		LeaderElectionID:   "apex-operator-lock",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconciler := &controller.Reconciler{
		Client:   mgr.GetClient(),
		Log:      mgr.GetLogger().WithValues("controller", "apex"),
		Scrapers: scraper.NewManager(),
		Metrics:  metrics,
	}

	err = reconciler.SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	err = (&apexv1.Scraper{}).SetupWebhookWithManager(mgr, certDir)
	if err != nil {
		setupLog.Error(err, "Unable to setup webhooks", "webhook", "Scraper")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
