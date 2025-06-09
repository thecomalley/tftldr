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
	// Try to get display_name or name, default to Unknown
	if displayName, ok := attributes["display_name"]; ok && displayName != nil {
		return fmt.Sprintf("%v", displayName)
	}
	if name, ok := attributes["name"]; ok && name != nil {
		return fmt.Sprintf("%v", name)
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

// Function to center a string in a specific width
func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}

	leftPadding := (width - len(s)) / 2
	rightPadding := width - len(s) - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
}

func displayTable(changes []ChangeRecord) {
	// Define color settings for each change type
	changeTypeColors := map[string]tablewriter.Colors{
		"create": tablewriter.Colors{tablewriter.FgGreenColor},
		"update": tablewriter.Colors{tablewriter.FgYellowColor},
		"delete": tablewriter.Colors{tablewriter.FgRedColor},
	}

	// Create a single table for all changes
	table := tablewriter.NewWriter(os.Stdout)

	// Create header with column names including Change Type
	table.SetHeader([]string{"Type", "Name", "Changed Parameters", "Resource Address"})

	// Set appearance options
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(true)
	table.SetTablePadding("\t")

	// Add rows for all changes
	for _, change := range changes {
		// Apply color to the change type column based on the change type
		typeColor := changeTypeColors[change.ChangeType]

		// Format change type as uppercase for better visibility
		changeTypeStr := strings.ToUpper(change.ChangeType)

		// Color settings for this row
		colors := []tablewriter.Colors{
			typeColor,
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
		}

		// Add row with rich text formatting
		table.Rich([]string{
			changeTypeStr,
			change.ResourceName,
			change.ChangedParams,
			change.ResourceAddress,
		}, colors)
	}

	table.Render()
}
