package model

type ClusterStatus struct {
	Instances          []InstanceState `json:"instanceStates"`
	Operators          []OperatorState `json:"operatorStates"`
	EnvironmentName    string          `json:"environmentName"`
	EnvironmentVersion string          `json:"environmentVersion"`
	PlanName           string          `json:"planName"`
	PlanVersion        string          `json:"planVersion"`
	TargetName         string          `json:"targetName"`
	TargetVersion      string          `json:"targetVersion"`
}

type InstanceState struct {
	Name            string `json:"name"`
	BlockID         string `json:"instanceId"`
	State           string `json:"state"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	DesiredReplicas int32  `json:"desiredReplicas"`
}

type OperatorState struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
