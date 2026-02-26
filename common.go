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
	rc.roomba.S.(io.Closer).Close()
	rc.roomba.Open(115200)
}

func (rc *roombaConn) StartHealthCheck() {
	rc.healthWorker = utils.NewStoppableWorkerWithTicker(2*time.Second, func(ctx context.Context) {
		// Get OI mode.
		_, err := rc.roomba.Sensors(35)
		if err != nil {
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

	return conn, nil
}

func releaseConn(serialPort string, logger logging.Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	conn, ok := connections[serialPort]
	if !ok {
		return
	}

	conn.refs--
	if conn.refs <= 0 {
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
