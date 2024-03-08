package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"gomess/cmd"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	execCtx, execCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	defer execCancel()

	go func() {
		<-execCtx.Done()
		logrus.Infof("received graceful shutdown signal")
	}()

	// command is expected to obey the cancellation signal on execCtx and
	// block while it is running
	if err := cmd.RootCommand().ExecuteContext(execCtx); err != nil {
		logrus.WithError(err).Fatal(err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// wait for API servers to shut down gracefully
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// wait for profiler, metrics and trace exporters to shut down gracefully
	}()

	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		wg.Wait()
	}()

	select {
	case <-shutdownCtx.Done():
		// cleanup timed out
		return

	case <-cleanupDone:
		// cleanup finished before timing out
		return
	}

}