package viamroomba

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/golang/geo/r3"
	base "go.viam.com/rdk/components/base"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/operation"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

var (
	Base             = resource.NewModel("dan", "viam-roomba", "base")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(base.API, Base,
		resource.Registration[base.Base, *Config]{
			Constructor: newViamRoombaBase,
		},
	)
}

type Config struct {
	SerialPort           string `json:"serial_port"`
	WidthMM              int    `json:"width_mm,omitempty"`
	WheelCircumferenceMM int    `json:"wheel_circumference_mm,omitempty"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.SerialPort == "" {
		return nil, nil, fmt.Errorf("%s: serial_port is required", path)
	}

	if cfg.WidthMM < 0 {
		return nil, nil, fmt.Errorf("%s: width_mm must be a positive number", path)
	}
	if cfg.WheelCircumferenceMM < 0 {
		return nil, nil, fmt.Errorf("%s: wheel_circumference_mm must be a positive number", path)
	}

	return nil, nil, nil
}

type viamRoombaBase struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	cfg    *Config

	conn       *roombaConn
	serialPort string

	widthMM              int
	wheelCircumferenceMM int

	opMgr       *operation.SingleOperationManager
	pos_x_mm    float64
	pos_y_mm    float64
	bearing_deg float64

	cancelCtx  context.Context
	cancelFunc func()
}

func newViamRoombaBase(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (base.Base, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewBase(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewBase(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (base.Base, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	conn, err := acquireConn(conf.SerialPort)
	if err != nil {
		logger.Infof("NewBase: failed to acquire connection on %s: %v", conf.SerialPort, err)
		cancelFunc()
		return nil, err
	}

	// Only enter Safe mode if the OI is currently off (mode == 0).
	// If it's already in Passive/Safe/Full, leave the current mode alone so
	// that a component rebuild (AlwaysRebuild) doesn't silently override a
	// mode the user intentionally set (e.g. Passive for charging).
	conn.mu.Lock()
	modeData, modeErr := conn.roomba.Sensors(35)
	logger.Infof("NewBase: initial OI mode sensor read, data=%v, err=%v", modeData, modeErr)
	if modeErr != nil || len(modeData) == 0 || modeData[0] == 0 {
		// OI is off (or unreadable) — send Safe to start it up.
		if err := conn.roomba.Safe(); err != nil {
			conn.mu.Unlock()
			cancelFunc()
			releaseConn(conf.SerialPort, logger)
			return nil, fmt.Errorf("failed to enter Safe mode: %w", err)
		}
		logger.Infof("NewBase: sent Safe() command on startup")
	} else {
		logger.Infof("NewBase: leaving existing OI mode as-is (mode=%d)", modeData[0])
	}
	conn.mu.Unlock()

	widthMM := conf.WidthMM
	if widthMM == 0 {
		widthMM = 235
	}
	wheelCircumferenceMM := conf.WheelCircumferenceMM
	if wheelCircumferenceMM == 0 {
		wheelCircumferenceMM = 220
	}

	s := &viamRoombaBase{
		name:                 name,
		logger:               logger,
		cfg:                  conf,
		conn:                 conn,
		serialPort:           conf.SerialPort,
		widthMM:              widthMM,
		wheelCircumferenceMM: wheelCircumferenceMM,
		opMgr:                operation.NewSingleOperationManager(),
		pos_x_mm:             0,
		pos_y_mm:             0,
		bearing_deg:          0,
		cancelCtx:            cancelCtx,
		cancelFunc:           cancelFunc,
	}

	logger.Infof("Roomba base initialized on %s (width: %dmm, wheel circumference: %dmm)",
		conf.SerialPort, widthMM, wheelCircumferenceMM)

	return s, nil
}

func (s *viamRoombaBase) Name() resource.Name {
	return s.name
}

// MoveStraight moves the robot straight a given distance at a given speed.
// If a distance or speed of zero is given, the base will stop.
// This method blocks until completed or cancelled.
func (s *viamRoombaBase) MoveStraight(ctx context.Context, distanceMm int, mmPerSec float64, extra map[string]any) error {
	ctx, done := s.opMgr.New(ctx)
	defer done()

	if distanceMm == 0 || mmPerSec == 0 {
		s.logger.Infof("MoveStraight: received zero distance or speed (distance=%d, speed=%.2f), stopping instead", distanceMm, mmPerSec)
		return s.Stop(ctx, extra)
	}

	duration := math.Abs(float64(distanceMm) / mmPerSec)

	var velocity int16
	if distanceMm > 0 {
		velocity = int16(mmPerSec)
	} else {
		velocity = -int16(mmPerSec)
	}

	if velocity > 500 {
		velocity = 500
	} else if velocity < -500 {
		velocity = -500
	}

	s.logger.Infof("MoveStraight: starting move (distance=%d mm, requested_speed=%.2f mm/sec, clamped_velocity=%d)", distanceMm, mmPerSec, velocity)

	s.conn.mu.Lock()
	if err := s.conn.roomba.Full(); err != nil {
		s.conn.mu.Unlock()
		return fmt.Errorf("failed to enter Full mode: %w", err)
	}
	if err := s.conn.roomba.Drive(velocity, 32767); err != nil {
		s.conn.mu.Unlock()
		return fmt.Errorf("failed to start straight movement: %w", err)
	}
	s.conn.mu.Unlock()

	s.logger.Infof("MoveStraight: distance=%d mm, velocity=%d mm/sec, duration=%.2f sec", distanceMm, velocity, duration)

	sleepCtx, cancel := context.WithTimeout(ctx, time.Duration(duration*1000)*time.Millisecond)
	defer cancel()

	select {
	case <-sleepCtx.Done():
		delta_x_mm := float64(distanceMm) * math.Cos(s.bearing_deg*math.Pi/180.0)
		delta_y_mm := float64(distanceMm) * math.Sin(s.bearing_deg*math.Pi/180.0)
		s.pos_x_mm += delta_x_mm
		s.pos_y_mm += delta_y_mm

		s.logger.Infof("Roomba Position: x=%.2f mm, y=%.2f mm, bearing=%.2f deg", s.pos_x_mm, s.pos_y_mm, s.bearing_deg)
	case <-ctx.Done():
		s.logger.Infof("MoveStraight: context cancelled, stopping move: %v", ctx.Err())
		s.Stop(ctx, extra)
		return ctx.Err()
	case <-s.cancelCtx.Done():
		s.logger.Infof("MoveStraight: base cancelCtx triggered, stopping move: %v", s.cancelCtx.Err())
		s.Stop(ctx, extra)
		return s.cancelCtx.Err()
	}

	return s.Stop(ctx, extra)
}

// Spin spins the robot by a given angle in degrees at a given speed.
// If a speed of 0 the base will stop.
// Given a positive speed and a positive angle, the base turns to the left (for built-in RDK drivers).
// This method blocks until completed or cancelled.
func (s *viamRoombaBase) Spin(ctx context.Context, angleDeg float64, degsPerSec float64, extra map[string]any) error {
	ctx, done := s.opMgr.New(ctx)
	defer done()

	if angleDeg == 0 || degsPerSec == 0 {
		s.logger.Infof("Spin: received zero angle or speed (angle=%.2f, speed=%.2f), stopping instead", angleDeg, degsPerSec)
		return s.Stop(ctx, extra)
	}

	// Convert angular speed (deg/s) to wheel velocity (mm/s).
	// For in-place spin: v = omega * (wheelbase / 2), omega in rad/s.
	wheelbaseMM := float64(s.widthMM)
	omegaRad := math.Abs(degsPerSec) * math.Pi / 180.0
	velocity := int16(omegaRad * wheelbaseMM / 2.0)
	if velocity > 500 {
		velocity = 500
	} else if velocity < 1 {
		velocity = 1
	}

	// Positive angleDeg → CCW (radius 1), negative → CW (radius -1).
	// Flip velocity sign for CW so the robot turns the right direction.
	var radius int16
	if angleDeg > 0 {
		radius = 1
	} else {
		radius = -1
	}

	duration := math.Abs(angleDeg / degsPerSec)

	s.logger.Infof("Spin: starting spin (angle=%.2f deg, speed=%.2f deg/sec, wheelbase=%.2f mm, computed_velocity=%d, radius=%d, duration=%.2f sec)",
		angleDeg, degsPerSec, wheelbaseMM, velocity, radius, duration)

	s.conn.mu.Lock()
	if err := s.conn.roomba.Full(); err != nil {
		s.conn.mu.Unlock()
		return fmt.Errorf("failed to enter Full mode: %w", err)
	}
	if err := s.conn.roomba.Drive(velocity, radius); err != nil {
		s.conn.mu.Unlock()
		return fmt.Errorf("failed to start spin: %w", err)
	}
	s.conn.mu.Unlock()

	s.logger.Infof("Spin: angle=%.2f deg, speed=%.2f deg/sec, velocity=%d mm/s, duration=%.2f sec", angleDeg, degsPerSec, velocity, duration)

	sleepCtx, cancel := context.WithTimeout(ctx, time.Duration(duration*1000)*time.Millisecond)
	defer cancel()

	select {
	case <-sleepCtx.Done():
		s.bearing_deg += angleDeg
		s.logger.Infof("Spin complete. Roomba Position: x=%.2f mm, y=%.2f mm, bearing=%.2f deg", s.pos_x_mm, s.pos_y_mm, s.bearing_deg)
	case <-ctx.Done():
		s.logger.Infof("Spin: context cancelled, stopping spin: %v", ctx.Err())
		s.Stop(ctx, extra)
		return ctx.Err()
	case <-s.cancelCtx.Done():
		s.logger.Infof("Spin: base cancelCtx triggered, stopping spin: %v", s.cancelCtx.Err())
		s.Stop(ctx, extra)
		return s.cancelCtx.Err()
	}

	return s.Stop(ctx, extra)
}

// SetPower sets the power of the base.
// For linear power, positive Y moves forwards for built-in RDK drivers.
// For angular power, positive Z turns to the left for built-in RDK drivers.
func (s *viamRoombaBase) SetPower(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]any) error {
	const maxWheelSpeed = 500.0
	maxAngularDegPerSec := maxWheelSpeed * 180.0 / (math.Pi * float64(s.widthMM) / 2.0)

	linearVel := r3.Vector{X: 0, Y: linear.Y * maxWheelSpeed, Z: 0}
	angularVel := r3.Vector{X: 0, Y: 0, Z: angular.Z * maxAngularDegPerSec}

	return s.SetVelocity(ctx, linearVel, angularVel, extra)
}

// SetVelocity sets the velocity of the base.
// linear is in mmPerSec (positive Y moves forwards for built-in RDK drivers).
// angular is in degsPerSec (positive Z turns to the left for built-in RDK drivers).
func (s *viamRoombaBase) SetVelocity(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]any) error {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	if linear.Y == 0 && angular.Z == 0 {
		s.logger.Infof("SetVelocity: zero linear and angular input, calling Stop()")
		return s.conn.roomba.Stop()
	}

	linearMM := linear.Y
	angularVel := angular.Z

	var velocity int16
	var radius int16

	if linearMM == 0 && angularVel != 0 {
		angularRadPerSec := math.Abs(angularVel) * math.Pi / 180.0
		wheelSpeed := angularRadPerSec * float64(s.widthMM) / 2.0
		velocity = int16(math.Min(500, wheelSpeed))
		if angularVel > 0 {
			radius = 1
		} else {
			radius = -1
		}
	} else {
		velocity = int16(linearMM)
		if velocity > 500 {
			s.logger.Warnf("Clamping velocity from %d to 500 mm/sec", velocity)
			velocity = 500
		} else if velocity < -500 {
			s.logger.Warnf("Clamping velocity from %d to -500 mm/sec", velocity)
			velocity = -500
		}

		if angularVel == 0 {
			radius = 32767 // Drive straight
		} else {
			radiusFloat := (float64(velocity) * 180.0) / (angularVel * math.Pi)
			radius = int16(math.Max(-2000, math.Min(2000, radiusFloat)))
		}
	}

	if err := s.conn.roomba.Full(); err != nil {
		return fmt.Errorf("failed to enter Full mode: %w", err)
	}
	if err := s.conn.roomba.Drive(velocity, radius); err != nil {
		return fmt.Errorf("failed to drive Roomba: %w", err)
	}

	s.logger.Infof("SetVelocity: linear_input=%.2f, angular_input=%.2f, velocity=%d mm/sec, radius=%d mm",
		linear.Y, angular.Z, velocity, radius)
	return nil
}

func (s *viamRoombaBase) Stop(ctx context.Context, extra map[string]any) error {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	if err := s.conn.roomba.Stop(); err != nil {
		return fmt.Errorf("failed to stop Roomba: %w", err)
	}

	s.logger.Info("Roomba stopped via Stop()")
	return nil
}

func (s *viamRoombaBase) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, fmt.Errorf("all commands have been moved to the brain service")
}

func (s *viamRoombaBase) IsMoving(ctx context.Context) (bool, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	// Packet 39: last requested velocity (0 after Stop(), non-zero while driving)
	data, err := s.conn.roomba.Sensors(39)
	if err != nil {
		s.logger.Infof("IsMoving: failed to read requested velocity sensor (packet 39): %v", err)
		return false, fmt.Errorf("failed to read requested velocity: %w", err)
	}
	if len(data) < 2 {
		s.logger.Infof("IsMoving: invalid sensor data length for packet 39 (len=%d)", len(data))
		return false, fmt.Errorf("invalid sensor data length")
	}

	requestedVelocity := int16(binary.BigEndian.Uint16(data))
	isMoving := math.Abs(float64(requestedVelocity)) > 5

	s.logger.Infof("IsMoving: requested_velocity=%d mm/s, moving=%v", requestedVelocity, isMoving)
	return isMoving, nil
}

// readAngle returns the angle in degrees since the last time it was requested.
// Packet ID 20, 2-byte signed integer.
func (s *viamRoombaBase) readAngle() (int16, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	s.conn.flushRx()
	data, err := s.conn.roomba.Sensors(20)
	if err != nil {
		s.logger.Infof("readAngle: sensor read failed: %v", err)
		return 0, err
	}
	if len(data) < 2 {
		s.logger.Infof("readAngle: unexpected sensor data length: %d", len(data))
		return 0, errors.New("unexpected sensor data length")
	}

	angle := int16(binary.BigEndian.Uint16(data))
	s.logger.Infof("readAngle: raw_angle=%d deg (since last read)", angle)
	return angle, nil
}

func (s *viamRoombaBase) readDistance() (int16, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	s.conn.flushRx()
	data, err := s.conn.roomba.Sensors(19)
	if err != nil {
		s.logger.Infof("readDistance: sensor read failed: %v", err)
		return 0, err
	}
	if len(data) < 2 {
		s.logger.Infof("readDistance: unexpected sensor data length: %d", len(data))
		return 0, errors.New("unexpected sensor data length")
	}

	dist := int16(binary.BigEndian.Uint16(data))
	s.logger.Infof("readDistance: raw_distance=%d mm (since last read)", dist)
	return dist, nil
}

// Properties returns the width, turning radius, and wheel circumference of the physical base in meters.
func (s *viamRoombaBase) Properties(ctx context.Context, extra map[string]any) (base.Properties, error) {
	return base.Properties{
		WidthMeters:              float64(s.widthMM) / 1000.0,
		TurningRadiusMeters:      0.0, // Differential drive can turn in place
		WheelCircumferenceMeters: float64(s.wheelCircumferenceMM) / 1000.0,
	}, nil
}

func (s *viamRoombaBase) Geometries(ctx context.Context, extra map[string]any) ([]spatialmath.Geometry, error) {
	// Roomba 650: 340mm diameter, 92mm height. Sphere approximation preserves the circular footprint.
	geom, err := spatialmath.NewSphere(spatialmath.NewZeroPose(), 170.0, s.name.Name)
	if err != nil {
		return nil, err
	}
	return []spatialmath.Geometry{geom}, nil
}

func (s *viamRoombaBase) Close(ctx context.Context) error {
	s.conn.mu.Lock()
	if err := s.conn.roomba.Stop(); err != nil {
		s.logger.Warnf("Failed to stop Roomba during close: %v", err)
	}
	s.conn.mu.Unlock()

	s.cancelFunc()
	releaseConn(s.serialPort, s.logger)

	s.logger.Info("Roomba base closed")
	return nil
}
