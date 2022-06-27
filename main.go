package main

import (
	"flag"
	"fmt"

	"github.com/acorn-io/traefik-hub-publisher/pkg/controller"
	"github.com/acorn-io/traefik-hub-publisher/pkg/version"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	versionFlag = flag.Bool("version", false, "print version")
)

func main() {
	flag.Parse()

	fmt.Printf("Version: %s\n", version.Get())
	if *versionFlag {
		return
	}

	ctx := signals.SetupSignalHandler()
	logrus.SetLevel(logrus.DebugLevel)
	if err := controller.Start(ctx); err != nil {
		logrus.Fatal(err)
	}
	<-ctx.Done()
	logrus.Fatal(ctx.Err())
}
