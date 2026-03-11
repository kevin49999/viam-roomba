package viamroomba

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	commonPB "go.viam.com/api/common/v1"
	pb "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	RoombaWorld = resource.NewModel("dan", "viam-roomba", "world")
)

var (
	roombaUUID  = "roomba-position"
	boxMetadata = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"color": {
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"r": {Kind: &structpb.Value_NumberValue{NumberValue: 0}},
							"g": {Kind: &structpb.Value_NumberValue{NumberValue: 255}},
							"b": {Kind: &structpb.Value_NumberValue{NumberValue: 0}},
						},
					},
				},
			},
			"opacity": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 0.5,
				},
			},
		},
	}
)

func init() {
	resource.RegisterService(worldstatestore.API, RoombaWorld,
		resource.Registration[worldstatestore.Service, *WorldConfig]{
			Constructor: newRoombaWorldRoombaWorld,
		},
	)
}

type WorldConfig struct {
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns three values:
//  1. Required dependencies: other resources that must exist for this resource to work.
//  2. Optional dependencies: other resources that may exist but are not required.
//  3. An error if any Config fields are missing or invalid.
//
// The `path` parameter indicates
// where this resource appears in the machine's JSON configuration
// (for example, "components.0"). You can use it in error messages
// to indicate which resource has a problem.
func (cfg *WorldConfig) Validate(path string) ([]string, []string, error) {
	// Add config validation code here
	return nil, nil, nil
}

type roombaWorldRoombaWorld struct {
	resource.AlwaysRebuild
	mu   sync.RWMutex
	name resource.Name

	logger    logging.Logger
	cfg       *WorldConfig
	obstacles map[string]*commonPB.Transform
	fps       float64

	changeChan chan worldstatestore.TransformChange
	cancelCtx  context.Context
	cancelFunc func()
}

func newRoombaWorldRoombaWorld(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (worldstatestore.Service, error) {
	conf, err := resource.NativeConfig[*WorldConfig](rawConf)
	if err != nil {
		return nil, err
	}

	return NewRoombaWorld(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewRoombaWorld(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *WorldConfig, logger logging.Logger) (worldstatestore.Service, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &roombaWorldRoombaWorld{
		name:       name,
		logger:     logger,
		cfg:        conf,
		obstacles:  make(map[string]*commonPB.Transform),
		changeChan: make(chan worldstatestore.TransformChange, 100),
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}

	// TEMP add dumby box
	s.obstacles[roombaUUID] = &commonPB.Transform{
		ReferenceFrame: "roomba-position",
		PoseInObserverFrame: &commonPB.PoseInFrame{
			ReferenceFrame: "world",
			Pose: &commonPB.Pose{
				X: 0, Y: 0, Z: 0, Theta: 0, OX: 0, OY: 0, OZ: 1,
			},
		},
		PhysicalObject: &commonPB.Geometry{
			GeometryType: &commonPB.Geometry_Box{
				Box: &commonPB.RectangularPrism{
					DimsMm: &commonPB.Vector3{
						X: 500,
						Y: 500,
						Z: 100,
					},
				},
			},
		},
		Uuid:     []byte(roombaUUID),
		Metadata: boxMetadata,
	}
	return s, nil
}

func (s *roombaWorldRoombaWorld) Name() resource.Name {
	return s.name
}

func (s *roombaWorldRoombaWorld) ListUUIDs(ctx context.Context, extra map[string]interface{}) ([][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uuids := make([][]byte, 0, len(s.obstacles))
	for _, obstacle := range s.obstacles {
		uuids = append(uuids, obstacle.Uuid)
	}

	return uuids, nil
}

func (s *roombaWorldRoombaWorld) GetTransform(ctx context.Context, uuid []byte, extra map[string]interface{}) (*commonPB.Transform, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transform, exists := s.obstacles[string(uuid)]
	if !exists {
		return nil, fmt.Errorf("transform not found")
	}

	return transform, nil
}

// StreamTransformChanges returns a channel of transform changes.
func (s *roombaWorldRoombaWorld) StreamTransformChanges(
	ctx context.Context,
	extra map[string]any,
) (*worldstatestore.TransformChangeStream, error) {
	return worldstatestore.NewTransformChangeStreamFromChannel(ctx, s.changeChan), nil
}

func (s *roombaWorldRoombaWorld) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	cmdName, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command must be a string")
	}

	switch cmdName {
	case "update_roomba_position":
		s.mu.Lock()
		x, ok := cmd["x"].(float64)
		if !ok {
			s.mu.Unlock()
			return nil, fmt.Errorf("x must be a number")
		}
		y, ok := cmd["y"].(float64)
		if !ok {
			s.mu.Unlock()
			return nil, fmt.Errorf("y must be a number")
		}

		roombaTransform := s.obstacles[roombaUUID]
		roombaTransform.PoseInObserverFrame.Pose.X = x * 1000
		roombaTransform.PoseInObserverFrame.Pose.Y = y * 1000
		s.mu.Unlock()
		s.emitTransformUpdate(&commonPB.Transform{
			PoseInObserverFrame: &commonPB.PoseInFrame{
				Pose: &commonPB.Pose{
					X: x * 1000, Y: y * 1000, Z: 0, Theta: 0, OX: 0, OY: 0, OZ: 1,
				},
			},
			Uuid: roombaTransform.Uuid,
		}, []string{"poseInObserverFrame.pose.x", "poseInObserverFrame.pose.y"})

		return map[string]interface{}{"status": "roomba_position_updated"}, nil
	case "draw_obstacle":
		x, ok := cmd["x"].(float64)
		if !ok {
			return nil, fmt.Errorf("x must be a number")
		}
		y, ok := cmd["y"].(float64)
		if !ok {
			return nil, fmt.Errorf("y must be a number")
		}
		label, ok := cmd["label"].(string)
		if !ok {
			return nil, fmt.Errorf("label must be a string")
		}

		uuid := make([]byte, 16)
		if _, err := rand.Read(uuid); err != nil {
			return nil, fmt.Errorf("failed to generate uuid: %w", err)
		}

		// x, y, z are in meters; Pose uses millimeters
		transform := &commonPB.Transform{
			ReferenceFrame: label,
			PoseInObserverFrame: &commonPB.PoseInFrame{
				ReferenceFrame: "world",
				Pose: &commonPB.Pose{
					X:  x * 1000,
					Y:  y * 1000,
					Z:  0,
					OZ: 1,
				},
			},
			PhysicalObject: &commonPB.Geometry{
				Center: &commonPB.Pose{X: 0, Y: 0, Z: 0, OZ: 1},
				GeometryType: &commonPB.Geometry_Box{
					Box: &commonPB.RectangularPrism{
						DimsMm: &commonPB.Vector3{X: 1000, Y: 1000, Z: 1000},
					},
				},
			},
			Uuid: uuid,
		}
		s.obstacles[string(uuid)] = transform
		s.logger.Infof("draw_obstacle: added obstacle at (%.2f, %.2f) m, total=%d", x, y, len(s.obstacles))
		s.emitTransformChange(transform, pb.TransformChangeType_TRANSFORM_CHANGE_TYPE_ADDED, nil)
		return map[string]interface{}{"status": "obstacle_added", "count": len(s.obstacles)}, nil

	case "clear_obstacles":
		s.mu.Lock()
		clearedCount := 0
		for obstacleUUID := range s.obstacles {
			if obstacleUUID != roombaUUID {
				delete(s.obstacles, obstacleUUID)
				clearedCount++
			}
		}
		s.logger.Infof("clear_obstacles: cleared %d obstacles", clearedCount)
		s.mu.Unlock()
		return map[string]interface{}{"status": "obstacles_cleared", "count": clearedCount}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}

func (s *roombaWorldRoombaWorld) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}

func (s *roombaWorldRoombaWorld) emitTransformChange(transform *commonPB.Transform, changeType pb.TransformChangeType, updatedFields []string) {
	change := worldstatestore.TransformChange{
		ChangeType:    changeType,
		Transform:     transform,
		UpdatedFields: updatedFields,
	}

	select {
	case s.changeChan <- change:
	case <-s.cancelCtx.Done():
	default:
		// Channel is full, skip this update
	}
}

func (s *roombaWorldRoombaWorld) emitTransformUpdate(partial *commonPB.Transform, updatedFields []string) {
	if partial == nil || len(partial.GetUuid()) == 0 {
		return
	}
	change := worldstatestore.TransformChange{
		ChangeType:    pb.TransformChangeType_TRANSFORM_CHANGE_TYPE_UPDATED,
		Transform:     partial,
		UpdatedFields: updatedFields,
	}
	select {
	case s.changeChan <- change:
	case <-s.cancelCtx.Done():
	default:
		// Channel is full, skip this update
	}
}
