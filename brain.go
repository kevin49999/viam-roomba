package viamroomba

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	rbase "go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	generic "go.viam.com/rdk/services/generic"
)

var (
	Brain = resource.NewModel("dan", "viam-roomba", "brain")
)

func init() {
	resource.RegisterService(generic.API, Brain,
		resource.Registration[resource.Resource, *BrainConfig]{
			Constructor: newBrainBrain,
		},
	)
}

type BrainConfig struct {
	BaseComponent    string `json:"base"`
	RoombaSensor     string `json:"roomba_sensor"`
	UltrasonicSensor string `json:"ultrasonic_sensor"`
}

func (cfg *BrainConfig) Validate(path string) ([]string, []string, error) {
	var required []string
	if cfg.BaseComponent != "" {
		required = append(required, cfg.BaseComponent)
	}
	if cfg.RoombaSensor != "" {
		required = append(required, cfg.RoombaSensor)
	}
	if cfg.UltrasonicSensor != "" {
		required = append(required, cfg.UltrasonicSensor)
	}
	return required, nil, nil
}

type ObstaclePosition struct {
	x_mm float64
	y_mm float64
}

type brainBrain struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	cfg    *BrainConfig

	base_            rbase.Base
	roombaSensor     sensor.Sensor
	ultrasonicSensor sensor.Sensor

	in_automatic_mode atomic.Bool
	obstaclePositions []ObstaclePosition
	pos_x_mm          float64
	pos_y_mm          float64
	bearing_deg       float64

	cancelCtx  context.Context
	cancelFunc func()
}

func newBrainBrain(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*BrainConfig](rawConf)
	if err != nil {
		return nil, err
	}

	return NewBrain(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewBrain(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *BrainConfig, logger logging.Logger) (resource.Resource, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &brainBrain{
		name:              name,
		logger:            logger,
		cfg:               conf,
		obstaclePositions: []ObstaclePosition{},
		cancelCtx:         cancelCtx,
		cancelFunc:        cancelFunc,
	}

	if conf.BaseComponent != "" {
		baseComp, err := rbase.FromProvider(deps, conf.BaseComponent)
		if err != nil {
			cancelFunc()
			return nil, fmt.Errorf("failed to get base component %q: %w", conf.BaseComponent, err)
		}
		s.base_ = baseComp
	}

	if conf.RoombaSensor != "" {
		roombaSensor, err := sensor.FromProvider(deps, conf.RoombaSensor)
		if err != nil {
			cancelFunc()
			return nil, fmt.Errorf("failed to get roomba sensor %q: %w", conf.RoombaSensor, err)
		}
		s.roombaSensor = roombaSensor
	}

	if conf.UltrasonicSensor != "" {
		ultrasonicSensor, err := sensor.FromProvider(deps, conf.UltrasonicSensor)
		if err != nil {
			cancelFunc()
			return nil, fmt.Errorf("failed to get ultrasonic sensor %q: %w", conf.UltrasonicSensor, err)
		}
		s.ultrasonicSensor = ultrasonicSensor
	}

	return s, nil
}

func (s *brainBrain) Name() resource.Name {
	return s.name
}

func (s *brainBrain) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	cmdName, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command must be a string")
	}

	switch cmdName {
	case "toggle_automatic_mode":
		if s.base_ == nil {
			return nil, fmt.Errorf("no base component configured, cannot toggle automatic mode")
		}
		if s.roombaSensor == nil || s.ultrasonicSensor == nil {
			return nil, fmt.Errorf("sensors not configured, cannot toggle automatic mode")
		}
		for {
			prev := s.in_automatic_mode.Load()
			if s.in_automatic_mode.CompareAndSwap(prev, !prev) {
				break
			}
		}
		s.logger.Infof("Automatic mode toggled to %v", s.in_automatic_mode.Load())
		if s.in_automatic_mode.Load() {
			go s.start_automatic_mode()
		} else {
			s.logger.Infof("Automatic mode stopped")
		}
		return map[string]any{"automatic_mode": s.in_automatic_mode.Load()}, nil

	case "get_automatic_mode":
		return map[string]any{"automatic_mode": s.in_automatic_mode.Load()}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}

func (s *brainBrain) Close(context.Context) error {
	s.cancelFunc()
	return nil
}

func (s *brainBrain) choose_direction() float64 {
	near_obstacles := []ObstaclePosition{}
	for _, obstacle := range s.obstaclePositions {
		distance := math.Sqrt(math.Pow(obstacle.x_mm-s.pos_x_mm, 2) + math.Pow(obstacle.y_mm-s.pos_y_mm, 2))
		if distance < 500 {
			near_obstacles = append(near_obstacles, obstacle)
		}
	}
	if len(near_obstacles) == 0 {
		return 90.0 + rand.Float64()*90.0 - 45.0
	}

	average_x := 0.0
	average_y := 0.0
	for _, obstacle := range near_obstacles {
		average_x += obstacle.x_mm
		average_y += obstacle.y_mm
	}
	average_x /= float64(len(near_obstacles))
	average_y /= float64(len(near_obstacles))

	direction := math.Atan2(average_y-s.pos_y_mm, average_x-s.pos_x_mm) * 180.0 / math.Pi
	return direction + 180.0 + rand.Float64()*90.0 - 45.0
}

func (s *brainBrain) add_obstacle_position(distance_in_front_of_roomba_mm float64) {
	obstacle_x := s.pos_x_mm + distance_in_front_of_roomba_mm*math.Cos(s.bearing_deg*math.Pi/180.0)
	obstacle_y := s.pos_y_mm + distance_in_front_of_roomba_mm*math.Sin(s.bearing_deg*math.Pi/180.0)
	s.logger.Infof("add_obstacle_position: obstacle_x=%f mm, obstacle_y=%f mm", obstacle_x, obstacle_y)
	s.obstaclePositions = append(s.obstaclePositions, ObstaclePosition{x_mm: obstacle_x, y_mm: obstacle_y})
}

// WARNING: very messy code below
func (s *brainBrain) start_automatic_mode() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.pos_x_mm = 0
	s.pos_y_mm = 0
	s.bearing_deg = 0
	s.obstaclePositions = []ObstaclePosition{}
	s.base_.DoCommand(ctx, map[string]any{"command": "play_song"})
	s.logger.Info("start_automatic_mode: entered")
	// TODO: the is_in_automatic_mode check should be in the loop but its being weird
	// NOTE: TO EXIT AUTO MODE WE HAVE TO RESTART THE MODULE ( there was a weird bug where it would reconfigure and then automatic mode would stop)
	for s.in_automatic_mode.Load() {
		s.logger.Info("start_automatic_mode: loop iteration starting")
		// move forward a bit
		distanceMm := 500
		mmPerSec := 500.0
		s.logger.Infof("start_automatic_mode: moving straight distance=%d mm, speed=%.1f mm/s", distanceMm, mmPerSec)
		if err := s.base_.MoveStraight(ctx, distanceMm, mmPerSec, nil); err != nil {
			s.logger.Infof("start_automatic_mode: MoveStraight failed: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		s.logger.Info("start_automatic_mode: move straight completed")
		// update position tracking
		delta_x_mm := float64(distanceMm) * math.Cos(s.bearing_deg*math.Pi/180.0)
		delta_y_mm := float64(distanceMm) * math.Sin(s.bearing_deg*math.Pi/180.0)
		s.pos_x_mm += delta_x_mm
		s.pos_y_mm += delta_y_mm
		// wait for distance/mmPerSec seconds
		sleepDur := time.Duration(float64(distanceMm)/mmPerSec) * time.Second
		s.logger.Infof("start_automatic_mode: sleeping %.2f s after move", sleepDur.Seconds())
		time.Sleep(sleepDur)
		s.logger.Info("start_automatic_mode: reading sensor readings")
		readings, err := s.roombaSensor.Readings(ctx, nil)
		if err != nil {
			s.logger.Infof("start_automatic_mode: Readings failed: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		s.logger.Info("start_automatic_mode: got sensor readings, checking obstacles")

		// Immediate check if we are actually going to hit something
		// Bumps are really bad wall and cliffs are probably fine (maybe split those up asp)
		bump_right := readings["bump_right"].(bool)
		bump_left := readings["bump_left"].(bool)
		wall := readings["wall"].(bool)
		cliff_left := readings["cliff_left"].(bool)
		cliff_front_left := readings["cliff_front_left"].(bool)
		cliff_front_right := readings["cliff_front_right"].(bool)
		cliff_right := readings["cliff_right"].(bool)

		if bump_right || bump_left || wall || cliff_left || cliff_front_left || cliff_front_right || cliff_right {
			s.add_obstacle_position(10)
			//  TODO: place obstacles in a world service here
			s.logger.Infof("start_automatic_mode: obstacles detected, turning away [bump_left=%v bump_right=%v wall=%v cliff_left=%v cliff_front_left=%v cliff_front_right=%v cliff_right=%v]",
				bump_left, bump_right, wall, cliff_left, cliff_front_left, cliff_front_right, cliff_right)
			// go backwards
			distanceMm := 500
			mmPerSec := 500.0
			s.logger.Infof("start_automatic_mode: moving backwards distance=%d mm, speed=%.1f mm/s", distanceMm, mmPerSec)
			if err := s.base_.MoveStraight(ctx, -distanceMm, mmPerSec, nil); err != nil {
				s.logger.Infof("start_automatic_mode: MoveStraight failed: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.logger.Info("start_automatic_mode: move backwards completed")
			delta_x_mm := float64(-distanceMm) * math.Cos(s.bearing_deg*math.Pi/180.0)
			delta_y_mm := float64(-distanceMm) * math.Sin(s.bearing_deg*math.Pi/180.0)
			s.pos_x_mm += delta_x_mm
			s.pos_y_mm += delta_y_mm
			angleDeg := s.choose_direction()
			degsPerSec := 100.0
			s.logger.Infof("start_automatic_mode: spinning angle=%.1f deg at %.1f deg/s", angleDeg, degsPerSec)
			if err := s.base_.Spin(ctx, angleDeg, degsPerSec, nil); err != nil {
				s.logger.Infof("start_automatic_mode: Spin failed: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.logger.Info("start_automatic_mode: spin completed")
			s.bearing_deg += angleDeg
			// wait for angleDeg/degsPerSec seconds
			spinSleepDur := time.Duration(math.Abs(angleDeg)/degsPerSec) * time.Second
			s.logger.Infof("start_automatic_mode: sleeping %.2f s after spin", spinSleepDur.Seconds())
			time.Sleep(spinSleepDur)
			s.logger.Info("start_automatic_mode: continuing loop after obstacle turn")
			continue
		}

		s.logger.Info("start_automatic_mode: no obstacles, continuing loop")
		timeoutContext, cancelTimeout := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer cancelTimeout()
		ultrasonic_readings, err := s.ultrasonicSensor.Readings(timeoutContext, nil)
		if err == context.DeadlineExceeded {
			s.logger.Infof("start_automatic_mode: ultrasonic sensor read timed out")
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			s.logger.Infof("start_automatic_mode: Readings failed: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		s.logger.Info("start_automatic_mode: got ultrasonic sensor readings, checking obstacles")
		ultrasonic_distance_m := ultrasonic_readings["distance"].(float64)
		s.logger.Infof("start_automatic_mode: ultrasonic_distance=%f m", ultrasonic_distance_m)
		if ultrasonic_distance_m < 0.75 {
			s.add_obstacle_position(ultrasonic_distance_m * 1000.0)
			s.logger.Infof("start_automatic_mode: ultrasonic_distance is too close, turning away")
			// go backwards
			distanceMm := -500
			mmPerSec := 500.0
			s.logger.Infof("start_automatic_mode: moving backwards distance=%d mm, speed=%.1f mm/s", distanceMm, mmPerSec)
			if err := s.base_.MoveStraight(ctx, distanceMm, mmPerSec, nil); err != nil {
				s.logger.Infof("start_automatic_mode: MoveStraight failed: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.logger.Info("start_automatic_mode: move backwards completed")
			delta_x_mm := float64(distanceMm) * math.Cos(s.bearing_deg*math.Pi/180.0)
			delta_y_mm := float64(distanceMm) * math.Sin(s.bearing_deg*math.Pi/180.0)
			s.pos_x_mm += delta_x_mm
			s.pos_y_mm += delta_y_mm
			angleDeg := s.choose_direction()
			degsPerSec := 100.0
			s.logger.Infof("start_automatic_mode: spinning angle=%.1f deg at %.1f deg/s", angleDeg, degsPerSec)
			if err := s.base_.Spin(ctx, angleDeg, degsPerSec, nil); err != nil {
				s.logger.Infof("start_automatic_mode: Spin failed: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.logger.Info("start_automatic_mode: spin completed")
			s.bearing_deg += angleDeg
			// wait for angleDeg/degsPerSec seconds
			spinSleepDur := time.Duration(math.Abs(angleDeg)/degsPerSec) * time.Second
			s.logger.Infof("start_automatic_mode: sleeping %.2f s after spin", spinSleepDur.Seconds())
			time.Sleep(spinSleepDur)
			s.logger.Info("start_automatic_mode: continuing loop after ultrasonic obstacle turn")
			continue
		}
		// TODO: use obstacle detect service to check if other stuff in way with non roomba native sensors (&& place obstacles in a world service here)

	}
	s.logger.Info("start_automatic_mode: exited (in_automatic_mode is false)")
	return nil
}

// IsInAutomaticMode returns true if the brain is currently in automatic mode.
// Used by the base to block manual movement commands.
func (s *brainBrain) IsInAutomaticMode() bool {
	return s.in_automatic_mode.Load()
}

