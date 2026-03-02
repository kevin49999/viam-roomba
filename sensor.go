package viamroomba

import (
	"context"
	"encoding/binary"
	"fmt"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var Sensor = resource.NewModel("jalen", "viam-roomba", "sensor")

func init() {
	resource.RegisterComponent(sensor.API, Sensor,
		resource.Registration[sensor.Sensor, *SensorConfig]{
			Constructor: newViamRoombaSensor,
		},
	)
}

type SensorConfig struct {
	SerialPort       string   `json:"serial_port"`
	RequestedSensors []string `json:"requested_sensors,omitempty"`
}

func (cfg *SensorConfig) Validate(path string) ([]string, []string, error) {
	if cfg.SerialPort == "" {
		return nil, nil, fmt.Errorf("%s: serial_port is required", path)
	}
	for _, name := range cfg.RequestedSensors {
		if _, ok := sensorToPackets[name]; !ok {
			return nil, nil, fmt.Errorf("%s: unknown sensor %q in requested_sensors", path, name)
		}
	}
	return nil, nil, nil
}

type viamRoombaSensor struct {
	resource.AlwaysRebuild

	name             resource.Name
	logger           logging.Logger
	conn             *roombaConn
	serialPort       string
	requestedSensors []string
}

func newViamRoombaSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*SensorConfig](rawConf)
	if err != nil {
		return nil, err
	}

	conn, err := acquireConn(conf.SerialPort)
	if err != nil {
		return nil, err
	}

	logger.Infof("Roomba sensor initialized on %s", conf.SerialPort)

	return &viamRoombaSensor{
		name:             rawConf.ResourceName(),
		logger:           logger,
		conn:             conn,
		serialPort:       conf.SerialPort,
		requestedSensors: conf.RequestedSensors,
	}, nil
}

func (s *viamRoombaSensor) Name() resource.Name {
	return s.name
}

// allSensorPackets lists all packet IDs in a stable order for querying.
var allSensorPackets = []byte{
	7,  // Bumps and Wheel Drops
	8,  // Wall
	9,  // Cliff Left
	10, // Cliff Front Left
	11, // Cliff Front Right
	12, // Cliff Right
	13, // Virtual Wall
	14, // Overcurrents
	15, // Dirt Detect
	17, // IR Opcode Omnidirectional Receiver
	18, // Buttons
	21, // Charging State
	22, // Voltage (mV)
	23, // Current (mA, signed)
	24, // Temperature (°C, signed)
	25, // Battery Charge (mAh)
	26, // Battery Capacity (mAh)
	27, // Wall Signal
	28, // Cliff Left Signal
	29, // Cliff Front Left Signal
	30, // Cliff Front Right Signal
	31, // Cliff Right Signal
	34, // Charging Sources Available
	35, // OI Mode
	39, // Requested Velocity (mm/s, signed)
	40, // Requested Radius (mm, signed)
	52, // IR Opcode Left Receiver
	53, // IR Opcode Right Receiver
	58, // Stasis
}

// sensorToPackets maps each reading name to the packet ID(s) required to compute it.
// battery_percent requires both packet 25 and 26.
var sensorToPackets = map[string][]byte{
	"bump_right":               {7},
	"bump_left":                {7},
	"wheel_drop_right":         {7},
	"wheel_drop_left":          {7},
	"wall":                     {8},
	"cliff_left":               {9},
	"cliff_front_left":         {10},
	"cliff_front_right":        {11},
	"cliff_right":              {12},
	"virtual_wall":             {13},
	"overcurrent_side_brush":   {14},
	"overcurrent_main_brush":   {14},
	"overcurrent_right_wheel":  {14},
	"overcurrent_left_wheel":   {14},
	"dirt_detect":              {15},
	"ir_opcode_omni":           {17},
	"button_clean":             {18},
	"button_spot":              {18},
	"button_dock":              {18},
	"button_minute":            {18},
	"button_hour":              {18},
	"button_day":               {18},
	"button_schedule":          {18},
	"button_clock":             {18},
	"charging_state":           {21},
	"voltage_mv":               {22},
	"current_ma":               {23},
	"temperature_c":            {24},
	"battery_charge_mah":       {25},
	"battery_capacity_mah":     {26},
	"battery_percent":          {25, 26},
	"wall_signal":              {27},
	"cliff_left_signal":        {28},
	"cliff_front_left_signal":  {29},
	"cliff_front_right_signal": {30},
	"cliff_right_signal":       {31},
	"charger_internal":         {34},
	"charger_homebase":         {34},
	"oi_mode":                  {35},
	"requested_velocity_mms":   {39},
	"requested_radius_mm":      {40},
	"ir_opcode_left":           {52},
	"ir_opcode_right":          {53},
	"stasis":                   {58},
}

var chargingStates = []string{"not_charging", "reconditioning", "full_charging", "trickle_charging", "waiting", "charging_fault"}
var oiModes = []string{"off", "passive", "safe", "full"}

func (s *viamRoombaSensor) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	s.conn.flushRx()

	// Determine which packets to query. When requested_sensors is set, only
	// query the packets needed for those sensors; otherwise query all packets.
	var packetsToQuery []byte
	var requestedSet map[string]bool

	if len(s.requestedSensors) == 0 {
		packetsToQuery = allSensorPackets
	} else {
		requestedSet = make(map[string]bool, len(s.requestedSensors))
		neededPackets := make(map[byte]bool)
		for _, name := range s.requestedSensors {
			requestedSet[name] = true
			for _, pid := range sensorToPackets[name] {
				neededPackets[pid] = true
			}
		}
		// Preserve stable ordering from allSensorPackets.
		for _, pid := range allSensorPackets {
			if neededPackets[pid] {
				packetsToQuery = append(packetsToQuery, pid)
			}
		}
	}

	data, err := s.conn.roomba.QueryList(packetsToQuery)
	if err != nil {
		s.logger.Infof("Sensor.Readings: QueryList failed: %v", err)
		return nil, fmt.Errorf("failed to query sensors: %w", err)
	}
	if len(data) != len(packetsToQuery) {
		s.logger.Infof("Sensor.Readings: unexpected sensor data count: got %d, want %d", len(data), len(packetsToQuery))
		return nil, fmt.Errorf("unexpected sensor data count: got %d, want %d", len(data), len(packetsToQuery))
	}

	// Build packet ID -> raw bytes map for O(1) lookup by packet ID.
	packetData := make(map[byte][]byte, len(packetsToQuery))
	for i, pid := range packetsToQuery {
		packetData[pid] = data[i]
	}

	b := func(pid byte) byte { return packetData[pid][0] }
	i16 := func(pid byte) int16 { return int16(binary.BigEndian.Uint16(packetData[pid])) }
	u16 := func(pid byte) uint16 { return binary.BigEndian.Uint16(packetData[pid]) }
	has := func(pid byte) bool { _, ok := packetData[pid]; return ok }

	readings := map[string]any{}

	s.logger.Infof("Sensor.Readings: successfully read %d packets from Roomba", len(data))

	// Packet 7: Bumps and Wheel Drops
	if has(7) {
		bumps := b(7)
		readings["bump_right"] = bumps&0x01 != 0
		readings["bump_left"] = bumps&0x02 != 0
		readings["wheel_drop_right"] = bumps&0x04 != 0
		readings["wheel_drop_left"] = bumps&0x08 != 0
	}

	// Packets 8-12: Proximity sensors
	if has(8) {
		readings["wall"] = b(8)&0x01 != 0
	}
	if has(9) {
		readings["cliff_left"] = b(9)&0x01 != 0
	}
	if has(10) {
		readings["cliff_front_left"] = b(10)&0x01 != 0
	}
	if has(11) {
		readings["cliff_front_right"] = b(11)&0x01 != 0
	}
	if has(12) {
		readings["cliff_right"] = b(12)&0x01 != 0
	}

	// Packet 13: Virtual Wall
	if has(13) {
		readings["virtual_wall"] = b(13)&0x01 != 0
	}

	// Packet 14: Overcurrents (bitmask: side brush, main brush, right wheel, left wheel)
	if has(14) {
		overcurrents := b(14)
		readings["overcurrent_side_brush"] = overcurrents&0x01 != 0
		readings["overcurrent_main_brush"] = overcurrents&0x04 != 0
		readings["overcurrent_right_wheel"] = overcurrents&0x08 != 0
		readings["overcurrent_left_wheel"] = overcurrents&0x10 != 0
	}

	// Packet 15: Dirt Detect
	if has(15) {
		readings["dirt_detect"] = int(b(15))
	}

	// Packet 17: IR Opcode Omnidirectional
	if has(17) {
		readings["ir_opcode_omni"] = int(b(17))
	}

	// Packet 18: Buttons
	if has(18) {
		buttons := b(18)
		readings["button_clean"] = buttons&0x01 != 0
		readings["button_spot"] = buttons&0x02 != 0
		readings["button_dock"] = buttons&0x04 != 0
		readings["button_minute"] = buttons&0x08 != 0
		readings["button_hour"] = buttons&0x10 != 0
		readings["button_day"] = buttons&0x20 != 0
		readings["button_schedule"] = buttons&0x40 != 0
		readings["button_clock"] = buttons&0x80 != 0
	}

	// Packet 21: Charging State
	if has(21) {
		chargingIdx := int(b(21))
		if chargingIdx < len(chargingStates) {
			readings["charging_state"] = chargingStates[chargingIdx]
		} else {
			readings["charging_state"] = "unknown"
			s.logger.Infof("Sensor.Readings: unknown charging_state index=%d", chargingIdx)
		}
	}

	// Packets 22-26: Battery
	if has(22) {
		readings["voltage_mv"] = int(u16(22))
	}
	if has(23) {
		readings["current_ma"] = int(i16(23))
	}
	if has(24) {
		readings["temperature_c"] = int(int8(b(24)))
	}
	if has(25) {
		readings["battery_charge_mah"] = int(u16(25))
	}
	if has(26) {
		capacity := int(u16(26))
		readings["battery_capacity_mah"] = capacity
		if has(25) {
			charge := int(u16(25))
			if capacity > 0 {
				readings["battery_percent"] = float64(charge) / float64(capacity) * 100.0
			} else {
				s.logger.Infof("Sensor.Readings: battery capacity reported as 0 (charge=%d)", charge)
			}
		}
	}

	// Packets 27-31: Signal strengths
	if has(27) {
		readings["wall_signal"] = int(u16(27))
	}
	if has(28) {
		readings["cliff_left_signal"] = int(u16(28))
	}
	if has(29) {
		readings["cliff_front_left_signal"] = int(u16(29))
	}
	if has(30) {
		readings["cliff_front_right_signal"] = int(u16(30))
	}
	if has(31) {
		readings["cliff_right_signal"] = int(u16(31))
	}

	// Packet 34: Charging Sources Available
	if has(34) {
		charger := b(34)
		readings["charger_internal"] = charger&0x01 != 0
		readings["charger_homebase"] = charger&0x02 != 0
	}

	// Packet 35: OI Mode
	if has(35) {
		modeIdx := int(b(35))
		if modeIdx < len(oiModes) {
			readings["oi_mode"] = oiModes[modeIdx]
		} else {
			readings["oi_mode"] = "unknown"
			s.logger.Infof("Sensor.Readings: unknown oi_mode index=%d", modeIdx)
		}
	}

	// Packets 39-40: Requested motion
	if has(39) {
		readings["requested_velocity_mms"] = int(i16(39))
	}
	if has(40) {
		readings["requested_radius_mm"] = int(i16(40))
	}

	// Packets 52-53: IR Opcodes left/right
	if has(52) {
		readings["ir_opcode_left"] = int(b(52))
	}
	if has(53) {
		readings["ir_opcode_right"] = int(b(53))
	}

	// Packet 58: Stasis
	if has(58) {
		readings["stasis"] = uint8(b(58))
	}

	// If specific sensors were requested, filter output to only those keys.
	if requestedSet != nil {
		for key := range readings {
			if !requestedSet[key] {
				delete(readings, key)
			}
		}
	}

	s.logger.Infof("Sensor.Readings summary: %d readings returned", len(readings))

	return readings, nil
}

func (s *viamRoombaSensor) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, nil
}

func (s *viamRoombaSensor) Close(ctx context.Context) error {
	releaseConn(s.serialPort, s.logger)
	return nil
}
