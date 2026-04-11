package main

import (
	"os"

	"github.com/firminochangani/audited/internal/common/logs"
)

func main() {
	logger := logs.New()
	svc := &Service{
		logger: logger,
	}
	if err := svc.Run(); err != nil {
		logger.Error("service exited with an error", "error", err)
		os.Exit(1)
	}
}
