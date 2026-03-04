package viamroomba

import (
	"context"
	"fmt"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

var (
	RoombaWorld = resource.NewModel("dan", "viam-roomba", "world")
)

func init() {
	resource.RegisterService(worldstatestore.API, RoombaWorld,
		resource.Registration[worldstatestore.Service, *WorldConfig]{
			Constructor: newRoombaWorldRoombaWorld,
		},
	)
}

type WorldConfig struct {
	/*
		Put config attributes here. There should be public/exported fields
		with a `json` parameter at the end of each attribute.

		Example config struct:
			type Config struct {
				Pin   string `json:"pin"`
				Board string `json:"board"`
				MinDeg *float64 `json:"min_angle_deg,omitempty"`
			}

		If your model does not need a config, replace *Config in the init
		function with resource.NoNativeConfig
	*/
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

	name resource.Name

	logger logging.Logger
	cfg    *WorldConfig

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
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *roombaWorldRoombaWorld) Name() resource.Name {
	return s.name
}

func (s *roombaWorldRoombaWorld) ListUUIDs(ctx context.Context, extra map[string]interface{}) ([][]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *roombaWorldRoombaWorld) GetTransform(ctx context.Context, uuid []byte, extra map[string]interface{}) (*commonPB.Transform, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *roombaWorldRoombaWorld) StreamTransformChanges(ctx context.Context, extra map[string]interface{}) (*worldstatestore.TransformChangeStream, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *roombaWorldRoombaWorld) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *roombaWorldRoombaWorld) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
