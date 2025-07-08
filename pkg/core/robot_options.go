package core

// RobotOption represents a configuration option for a Robot
type RobotOption func(*Robot)

// WithName sets the robot's name
func WithName(name string) RobotOption {
	return func(r *Robot) {
		r.Name = name
	}
}

// WithWork sets the robot's work function
func WithWork(work func()) RobotOption {
	return func(r *Robot) {
		r.Work = work
	}
}

// WithAutoRun sets whether the robot should automatically run when started
func WithAutoRun(autoRun bool) RobotOption {
	return func(r *Robot) {
		r.AutoRun = autoRun
	}
}

// WithConnections adds connections to the robot
func WithConnections(connections ...Connection) RobotOption {
	return func(r *Robot) {
		for _, connection := range connections {
			r.AddConnection(connection)
		}
	}
}

// WithDevices adds devices to the robot
func WithDevices(devices ...Device) RobotOption {
	return func(r *Robot) {
		for _, device := range devices {
			r.AddDevice(device)
		}
	}
}