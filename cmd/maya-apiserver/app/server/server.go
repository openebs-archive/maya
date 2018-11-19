package server

import (
	"io"
	"log"
	"sync"

	"github.com/openebs/maya/cmd/maya-apiserver/app/config"
)

// MayaApiServer is a long running stateless daemon that runs
// at openebs maya master(s)
type MayaApiServer struct {
	config    *config.MayaConfig
	logger    *log.Logger
	logOutput io.Writer

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

// NewMayaApiServer is used to create a new maya api server
// with the given configuration
func NewMayaApiServer(config *config.MayaConfig, logOutput io.Writer) (*MayaApiServer, error) {
	ms := &MayaApiServer{
		config:     config,
		logger:     log.New(logOutput, "", log.LstdFlags|log.Lmicroseconds),
		logOutput:  logOutput,
		shutdownCh: make(chan struct{}),
	}
	return ms, nil
}

// Shutdown is used to terminate MayaServer.
func (ms *MayaApiServer) Shutdown() error {

	ms.shutdownLock.Lock()
	defer ms.shutdownLock.Unlock()

	ms.logger.Println("[INFO] maya api server: requesting shutdown")

	if ms.shutdown {
		return nil
	}

	ms.logger.Println("[INFO] maya api server: shutdown complete")
	ms.shutdown = true

	close(ms.shutdownCh)

	return nil
}

// Leave is used gracefully exit.
func (ms *MayaApiServer) Leave() error {

	ms.logger.Println("[INFO] maya api server: exiting gracefully")

	// Nothing as of now
	return nil
}
