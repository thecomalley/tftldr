package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

func main() {
	// Define command line flags
	inputFile := flag.String("input", "tfplan.json", "Path to the Terraform plan JSON file")
	csvOutput := flag.String("csv", "", "Path to export CSV (optional, for ITIL change tickets)")
	flag.Parse()

	// Load configuration
	loadConfig()

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

	// Export to CSV if requested
	if *csvOutput != "" {
		if err := exportToCSV(changes, *csvOutput); err != nil {
			fmt.Printf("Error exporting to CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported changes to CSV: %s\n", *csvOutput)
	}
}

func processChanges(resources []ResourceChange) []ChangeRecord {
	var changes []ChangeRecord

	for _, resource := range resources {
		// Skip ignored resource types
		if shouldIgnoreResource(resource.Type) {
			continue
		}

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

// shouldIgnoreResource determines if a resource type should be excluded from the output
func shouldIgnoreResource(resourceType string) bool {
	// Get ignored resource type prefixes from config
	ignoredPrefixes := viper.GetStringSlice("ignore.prefixes")

	// Get ignored exact resource types from config
	ignoredTypes := viper.GetStringSlice("ignore.types")

	// Check if it's an exact match with ignored types
	for _, t := range ignoredTypes {
		if resourceType == t {
			return true
		}
	}

	// Check if it starts with any of the ignored prefixes
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(resourceType, prefix) {
			return true
		}
	}

	return false
}

func getResourceName(attributes map[string]interface{}) string {
	// Try to get display_name, then name, then id, returning first one found
	if displayName, ok := attributes["display_name"]; ok && displayName != nil {
		return fmt.Sprint(displayName)
	}
	if name, ok := attributes["name"]; ok && name != nil {
		return fmt.Sprint(name)
	}
	if id, ok := attributes["id"]; ok && id != nil {
		return fmt.Sprint(id)
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
	// Define color settings for each change type
	changeTypeColors := map[string]tablewriter.Colors{
		"create": tablewriter.Colors{tablewriter.FgGreenColor},
		"update": tablewriter.Colors{tablewriter.FgYellowColor},
		"delete": tablewriter.Colors{tablewriter.FgRedColor},
	}

	// Create a single table for all changes
	table := tablewriter.NewWriter(os.Stdout)

	// Get column visibility settings from config
	showChangeType := viper.GetBool("columns.changeType")
	showResourceName := viper.GetBool("columns.resourceName")
	showChangedParams := viper.GetBool("columns.changedParams")
	showResourceType := viper.GetBool("columns.resourceType")
	showResourceAddress := viper.GetBool("columns.resourceAddress")

	// Build headers and column indices based on visibility settings
	var headers []string
	var columnIndices []int

	// Add columns in order based on visibility settings
	if showChangeType {
		headers = append(headers, "Type")
		columnIndices = append(columnIndices, 0)
	}
	if showResourceName {
		headers = append(headers, "Name")
		columnIndices = append(columnIndices, 1)
	}
	if showChangedParams {
		headers = append(headers, "Changed Parameters")
		columnIndices = append(columnIndices, 2)
	}
	if showResourceType {
		headers = append(headers, "Resource Type")
		columnIndices = append(columnIndices, 3)
	}
	if showResourceAddress {
		headers = append(headers, "Resource Address")
		columnIndices = append(columnIndices, 4)
	}

	// Create header with selected column names
	table.SetHeader(headers)

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
		// Format change type as uppercase for better visibility
		changeTypeStr := strings.ToUpper(change.ChangeType)

		// Prepare all possible values
		allValues := []string{
			changeTypeStr,
			change.ResourceName,
			change.ChangedParams,
			change.ResourceType,
			change.ResourceAddress,
		}

		// Select only the visible columns
		var rowValues []string
		for _, idx := range columnIndices {
			rowValues = append(rowValues, allValues[idx])
		}

		// Color settings for this row
		colors := make([]tablewriter.Colors, len(headers))

		// Apply color to the change type column if it's visible
		if showChangeType {
			colors[0] = changeTypeColors[change.ChangeType]
		}

		// Add row with rich text formatting
		table.Rich(rowValues, colors)
	}

	table.Render()
}

// exportToCSV writes the changes to a CSV file for ITIL change tickets
func exportToCSV(changes []ChangeRecord, outputPath string) error {
	// Create or truncate the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get column visibility settings from config
	showChangeType := viper.GetBool("columns.changeType")
	showResourceName := viper.GetBool("columns.resourceName")
	showChangedParams := viper.GetBool("columns.changedParams")
	showResourceType := viper.GetBool("columns.resourceType")
	showResourceAddress := viper.GetBool("columns.resourceAddress")

	// Build headers and column indices based on visibility settings
	var headers []string
	var columnIndices []int

	// Add columns in order based on visibility settings
	if showChangeType {
		headers = append(headers, "Change Type")
		columnIndices = append(columnIndices, 0)
	}
	if showResourceName {
		headers = append(headers, "Resource Name")
		columnIndices = append(columnIndices, 1)
	}
	if showChangedParams {
		headers = append(headers, "Changed Parameters")
		columnIndices = append(columnIndices, 2)
	}
	if showResourceType {
		headers = append(headers, "Resource Type")
		columnIndices = append(columnIndices, 3)
	}
	if showResourceAddress {
		headers = append(headers, "Resource Address")
		columnIndices = append(columnIndices, 4)
	}

	// Write header row
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write change records
	for _, change := range changes {
		// Prepare all possible values
		allValues := []string{
			strings.ToUpper(change.ChangeType),
			change.ResourceName,
			change.ChangedParams,
			change.ResourceType,
			change.ResourceAddress,
		}

		// Select only the visible columns
		var record []string
		for _, idx := range columnIndices {
			record = append(record, allValues[idx])
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}
