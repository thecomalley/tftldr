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
	configFile := flag.String("config", "", "Path to the configuration file (optional)")
	csvOutput := flag.String("csv", "", "Path to export CSV (optional, for ITIL change tickets)")
	flag.Parse()

	// Load configuration
	loadConfig(*configFile)

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

// loadConfig moved to config.go

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

	// Write header row
	header := []string{"Change Type", "Resource Name", "Changed Parameters", "Resource Type", "Terraform Resource Address"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write change records
	for _, change := range changes {
		record := []string{
			strings.ToUpper(change.ChangeType),
			change.ResourceName,
			change.ChangedParams,
			change.ResourceType,
			change.ResourceAddress,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}
