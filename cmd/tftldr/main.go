package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func main() {
	// Define command line flags
	inputFile := flag.String("input", "tfplan.json", "Path to the Terraform plan JSON file")
	flag.Parse()

	// Load the Terraform plan JSON file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var plan TerraformPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Process the changes
	changes := processChanges(plan.ResourceChanges)

	// Sort changes by change type
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].ChangeType < changes[j].ChangeType
	})

	// Display the table
	displayTable(changes)
}

func processChanges(resources []ResourceChange) []ChangeRecord {
	var changes []ChangeRecord

	for _, resource := range resources {
		for _, action := range resource.Change.Actions {
			// Skip no-op and read changes
			if action == "no-op" || action == "read" {
				continue
			}

			// Determine the source of data and parameters based on action
			var resourceName string
			var changedParams []string

			// For delete operations, get info from before state
			// For create/update operations, get info from after state
			attributes := resource.Change.Before
			if action == "create" || action == "update" {
				attributes = resource.Change.After
			}

			resourceName = getResourceName(attributes)

			// Only diff params for updates
			if action == "update" {
				changedParams = diffParams(resource.Change.Before, resource.Change.After)
			} else {
				// For create and delete, all params are changed
				changedParams = []string{"All parameters"}
			}

			changes = append(changes, ChangeRecord{
				ChangeType:      action,
				ResourceName:    resourceName,
				ChangedParams:   strings.Join(changedParams, ", "),
				ResourceType:    resource.Type,
				ResourceAddress: resource.Address,
			})
		}
	}

	return changes
}

func getResourceName(attributes map[string]interface{}) string {
	// Try to get name or display_name, default to Unknown
	if name, ok := attributes["name"]; ok && name != nil {
		return fmt.Sprintf("%v", name)
	}
	if displayName, ok := attributes["display_name"]; ok && displayName != nil {
		return fmt.Sprintf("%v", displayName)
	}
	return "Unknown"
}

func diffParams(before, after map[string]interface{}) []string {
	var changedParams []string

	for k, vAfter := range after {
		vBefore, exists := before[k]
		if !exists {
			changedParams = append(changedParams, k)
			continue
		}

		// Use reflect.DeepEqual for proper comparison of complex types
		if !reflect.DeepEqual(vBefore, vAfter) {
			changedParams = append(changedParams, k)
		}
	}

	// Also check for keys in before that don't exist in after
	for k := range before {
		if _, exists := after[k]; !exists {
			changedParams = append(changedParams, k)
		}
	}

	return changedParams
}

func displayTable(changes []ChangeRecord) {
	// Group changes by type
	changesByType := map[string][]ChangeRecord{
		"create": {},
		"update": {},
		"delete": {},
	}

	for _, change := range changes {
		if _, ok := changesByType[change.ChangeType]; ok {
			changesByType[change.ChangeType] = append(changesByType[change.ChangeType], change)
		}
	}

	// Define table appearance settings for each change type
	tableSettings := map[string]struct {
		title string
		color tablewriter.Colors
	}{
		"create": {"CREATE", tablewriter.Colors{tablewriter.FgGreenColor}},
		"update": {"UPDATE", tablewriter.Colors{tablewriter.FgYellowColor}},
		"delete": {"DELETE", tablewriter.Colors{tablewriter.FgRedColor}},
	}

	// Display tables for each change type
	for changeType, records := range changesByType {
		if len(records) == 0 {
			continue
		}

		settings := tableSettings[changeType]

		table := tablewriter.NewWriter(os.Stdout)

		// Create a multi-row header with the change type as the title
		table.SetHeader([]string{"Name", "Address", "Changed Parameters"})

		// Set appearance options
		table.SetColumnColor(
			settings.color,
			tablewriter.Colors{},
			tablewriter.Colors{},
		)

		for _, change := range records {
			table.Append([]string{
				change.ResourceName,
				change.ResourceAddress,
				change.ChangedParams,
			})
		}

		table.Render()
	}
}
