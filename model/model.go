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
	Name            string            `json:"name"`
	BlockID         string            `json:"instanceId"`
	State           string            `json:"state"`
	Metadata        map[string]string `json:"metadata"`
	ReadyReplicas   int32             `json:"readyReplicas"`
	DesiredReplicas int32             `json:"desiredReplicas"`
}

type OperatorState struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type TraefikRoutes []struct {
	EntryPoints []string `json:"entryPoints"`
	Service     string   `json:"service"`
	Rule        string   `json:"rule"`
	Priority    int64    `json:"priority,omitempty"`
	Status      string   `json:"status"`
	Using       []string `json:"using"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	Middlewares []string `json:"middlewares,omitempty"`
	TLS         TLS      `json:"tls,omitempty"`
}
type TLS struct {
	Options      string `json:"options"`
	CertResolver string `json:"certResolver"`
}

type TraefikService struct {
	LoadBalancer LoadBalancer      `json:"loadBalancer"`
	Status       string            `json:"status"`
	UsedBy       []string          `json:"usedBy"`
	ServerStatus map[string]string `json:"serverStatus"`
	Name         string            `json:"name"`
	Provider     string            `json:"provider"`
	Type         string            `json:"type"`
}
type Servers struct {
	URL string `json:"url"`
}
type LoadBalancer struct {
	Servers        []Servers `json:"servers"`
	PassHostHeader bool      `json:"passHostHeader"`
}
