package plan

const (
	RiskSafe        = "safe"
	RiskMigration   = "migration"
	RiskDestructive = "destructive"
)

type Plan struct {
	Summary    Summary     `json:"summary"`
	Operations []Operation `json:"operations"`
}

type Summary struct {
	Supported      bool `json:"supported"`
	Destructive    bool `json:"destructive"`
	OperationCount int  `json:"operationCount"`
}

type Operation struct {
	Kind       string `json:"kind"`
	ObjectType string `json:"objectType"`
	ObjectKey  string `json:"objectKey"`
	Risk       string `json:"risk"`
	Reason     string `json:"reason,omitempty"`
	SQL        string `json:"sql"`
}
