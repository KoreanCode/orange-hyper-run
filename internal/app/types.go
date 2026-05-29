package app

type planResult struct {
	Body    string
	Created bool
}

type projectState struct {
	Project          string `json:"project"`
	Stage            string `json:"stage"`
	Status           string `json:"status"`
	ActiveRunID      string `json:"active_run_id"`
	CurrentGoalID    string `json:"current_goal_id"`
	CurrentGoalPath  string `json:"current_goal_path"`
	ExecutionAdapter string `json:"execution_adapter"`
	PlanPath         string `json:"plan_path"`
	PlanHash         string `json:"plan_hash"`
	Focus            string `json:"focus"`
	AutoContinue     bool   `json:"auto_continue,omitempty"`
	RunUntil         string `json:"run_until,omitempty"`
	RunTargetSource  string `json:"run_target_source,omitempty"`
	UpdatedAt        string `json:"updated_at"`
}

type runOptions struct {
	Focus           string
	AutoContinue    bool
	RunUntil        string
	RunTargetSource string
}

type episode struct {
	Plan          map[string]string
	Stage         string
	BuildStyle    string
	Objective     string
	Scope         string
	NonGoals      string
	Validation    string
	StopCondition string
	Docs          episodeDocs
}

type episodeDocs struct {
	Goal     string
	Tasks    string
	Evidence string
	Review   string
	Next     string
}

type handoff struct {
	Adapter           string
	EventType         string
	Description       string
	InstructionsLabel string
	Instructions      string
}

type goalState struct {
	State  string
	Reason string
}

type memory struct {
	Kind       string
	Text       string
	Confidence float64
	Quality    string
}

type learnResult struct {
	Skipped     bool
	Reason      string
	State       string
	RunID       string
	GoalID      string
	Inserted    int
	MemoryCount int
	Quality     map[string]int
	Rejected    map[string]int
}

type similarContext struct {
	Source string
	ID     string
	Kind   string
	Text   string
	Score  float64
}

type growthState struct {
	Version         int               `json:"version"`
	UpdatedAt       string            `json:"updated_at"`
	PressureLedger  pressureLedger    `json:"pressure_ledger"`
	Pressures       []growthPressure  `json:"pressures"`
	RuntimeBehavior growthBehavior    `json:"runtime_behavior"`
	Candidates      []growthCandidate `json:"candidates"`
	Thresholds      growthThresholds  `json:"thresholds"`
}

type pressureLedger struct {
	Method              string   `json:"method"`
	Protocol            string   `json:"protocol"`
	Principles          []string `json:"principles"`
	OpenPressures       int      `json:"open_pressures"`
	CandidateStructures int      `json:"candidate_structures"`
	ActiveStructures    int      `json:"active_structures"`
}

type growthPressure struct {
	Kind            string   `json:"kind"`
	PressureType    string   `json:"pressure_type"`
	Signal          string   `json:"signal"`
	CanonicalSignal string   `json:"canonical_signal"`
	Effect          string   `json:"effect"`
	State           string   `json:"state"`
	GoalCount       int      `json:"goal_count"`
	MemoryCount     int      `json:"memory_count"`
	Score           float64  `json:"score"`
	Sources         []string `json:"sources"`
}

type growthBehavior struct {
	WorkBoundary      []string `json:"work_boundary"`
	ValidationSignals []string `json:"validation_signals"`
	StopConditions    []string `json:"stop_conditions"`
}

type growthCandidate struct {
	Kind                string   `json:"kind"`
	Name                string   `json:"name"`
	Status              string   `json:"status"`
	GeneratedPath       string   `json:"generated_path"`
	LifecyclePath       string   `json:"lifecycle_path"`
	Reason              string   `json:"reason"`
	Signal              string   `json:"signal"`
	PressureType        string   `json:"pressure_type"`
	Sources             []string `json:"sources"`
	EvidenceCount       int      `json:"evidence_count"`
	RepeatedThreshold   int      `json:"repeated_threshold"`
	PromotionThreshold  int      `json:"promotion_threshold"`
	ActivationThreshold int      `json:"activation_threshold"`
}

type growthThresholds struct {
	RepeatedSignalGoals      int `json:"repeated_signal_goals"`
	PromotableSignalGoals    int `json:"promotable_signal_goals"`
	ActiveSignalGoals        int `json:"active_signal_goals"`
	HarnessStablePressures   int `json:"harness_stable_pressures"`
	HarnessPromotableSignals int `json:"harness_promotable_signals"`
	HarnessActiveSignals     int `json:"harness_active_signals"`
}

type readinessState struct {
	Version      int                  `json:"version"`
	UpdatedAt    string               `json:"updated_at"`
	Stage        string               `json:"stage"`
	Dimensions   []readinessDimension `json:"dimensions"`
	StageGate    readinessStageGate   `json:"stage_gate"`
	NextPressure readinessPressure    `json:"next_pressure"`
}

type readinessDimension struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Score    int    `json:"score"`
	Evidence string `json:"evidence"`
	Gap      string `json:"gap"`
}

type readinessStageGate struct {
	CurrentStage     string                 `json:"current_stage"`
	NextStage        string                 `json:"next_stage"`
	Status           string                 `json:"status"`
	RequiredAxes     []string               `json:"required_axes"`
	BlockingGaps     []string               `json:"blocking_gaps"`
	RequiredEvidence []string               `json:"required_evidence"`
	Advancement      stageAdvancementPolicy `json:"advancement"`
}

type stageAdvancementPolicy struct {
	Candidate        bool     `json:"candidate"`
	Recommendation   string   `json:"recommendation"`
	PlanChange       string   `json:"plan_change"`
	RequiredEvidence []string `json:"required_evidence"`
}

type readinessPressure struct {
	Axis             string `json:"axis"`
	AxisName         string `json:"axis_name"`
	Status           string `json:"status"`
	Reason           string `json:"reason"`
	RecommendedGoal  string `json:"recommended_goal"`
	WorkBoundary     string `json:"work_boundary"`
	ValidationSignal string `json:"validation_signal"`
}
