package main

// ResourceChange represents a Terraform resource change
type ResourceChange struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Change  struct {
		Actions []string               `json:"actions"`
		Before  map[string]interface{} `json:"before"`
		After   map[string]interface{} `json:"after"`
	} `json:"change"`
}

// TerraformPlan represents the overall Terraform plan
type TerraformPlan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

// ChangeRecord represents a single change for output
type ChangeRecord struct {
	ChangeType      string
	ResourceName    string
	ChangedParams   string
	ResourceType    string
	ResourceAddress string
}
