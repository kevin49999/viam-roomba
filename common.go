package viamroomba

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/parabolala/go-roomba"
	"go.viam.com/rdk/logging"
	"go.viam.com/utils"
)

type roombaConn struct {
	roomba *roomba.Roomba
	mu     sync.Mutex
	refs   int

	healthWorker *utils.StoppableWorkers
}

// ResetConn assumes the `rc.mu` is held.
func (rc *roombaConn) ResetConn() {
	// Caller should already hold rc.mu; log before and after reset for debugging.
	start := time.Now()
	fmt.Println("roombaConn.ResetConn: resetting underlying serial connection")
	rc.roomba.S.(io.Closer).Close()
	rc.roomba.Open(115200)
	fmt.Printf("roombaConn.ResetConn: reset complete in %s\n", time.Since(start))
}

func (rc *roombaConn) StartHealthCheck() {
	rc.healthWorker = utils.NewStoppableWorkerWithTicker(2*time.Second, func(ctx context.Context) {
		// Get OI mode.
		_, err := rc.roomba.Sensors(35)
		if err != nil {
			fmt.Printf("roombaConn.StartHealthCheck: OI mode sensor read failed, resetting connection: %v\n", err)
			rc.ResetConn()
		}
	})
}

var (
	globalMu    sync.Mutex
	connections = map[string]*roombaConn{}
)

func acquireConn(serialPort string) (*roombaConn, error) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if conn, ok := connections[serialPort]; ok {
		conn.refs++
		fmt.Printf("acquireConn: reusing existing connection on %s (refs=%d)\n", serialPort, conn.refs)
		return conn, nil
	}
	r, err := roomba.MakeRoomba(serialPort)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial connection on %s: %w", serialPort, err)
	}
	// Send START command (opcode 128) to enable the Open Interface before any queries or commands.
	if err := r.Passive(); err != nil {
		return nil, fmt.Errorf("failed to start OI on %s: %w", serialPort, err)
	}
	conn := &roombaConn{roomba: r, refs: 1}
	conn.setReadTimeout(2 * time.Second)
	go conn.StartHealthCheck()
	connections[serialPort] = conn

	fmt.Printf("acquireConn: created new connection on %s\n", serialPort)

	return conn, nil
}

func releaseConn(serialPort string, logger logging.Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	conn, ok := connections[serialPort]
	if !ok {
		logger.Infof("releaseConn: attempted to release unknown connection on %s", serialPort)
		return
	}

	conn.refs--
	if conn.refs <= 0 {
		logger.Infof("releaseConn: closing connection on %s", serialPort)
		conn.healthWorker.Stop()
		delete(connections, serialPort)
		if closer, ok := conn.roomba.S.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				logger.Error("Closing serial connection errored:", err)
				panic("error closing serial connection")
			}
		} else {
			logger.Errorf("Serial connection was not a closer: %T", conn.roomba.S)
			panic("Could not close serial connection")
		}
	}
}
