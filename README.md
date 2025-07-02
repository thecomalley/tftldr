# tftldr 🔍

> "Because life's too short to read entire Terraform plans!"

## What is this sorcery? ✨

**tftldr** is your new best friend for making sense of those ridiculously verbose Terraform plan JSON files.

Ever stared at a Terraform plan that's longer than a fantasy novel? Ever wished you could just get the CliffsNotes version? Say no more! **tftldr** transforms that wall of JSON into a beautiful, color-coded table that even your project manager could understand.

![Terraform Plan Summary](docs/image.png)

## Features 🚀

- Turns intimidating JSON blobs into friendly tables
- Color codes changes (green for creations, yellow for updates, red for deletions)
- Summarizes what's actually changing without the fluff
- Filters out noise from utility resources like random providers and null resources
- Export to CSV for ITIL change tickets and documentation

## Seriously, why? 🤷‍♂️

Ok seriously though, there are a few terraform plan summary tools out there, but they focus on the terraform resource rather than the actual deployed resource. This is fine for those cattle scenarios, but in my experience theres still a lot of pets out there that need to be cared for. This tool is designed for Day2 operations when change management needs to be appeased before you can run your pipeline.


## Installation 📦

```bash
go install github.com/thecomalley/tftldr@latest
```

## Usage 🛠️

Generate a Terraform plan JSON file using:

```bash
terraform plan -out=tfplan.out
terraform show -json tfplan.out > tfplan.json
```

Run against your default `tfplan.json` file:

```bash
tftldr
```

Or specify a different plan file:

```bash
tftldr -input path/to/your/plan.json
```

Export to CSV for ITIL change tickets:

```bash
tftldr -csv changes.csv
```

You can combine multiple flags:

```bash
tftldr -input plan.json -config custom-config.yml -csv changes.csv
```

## Configuration 🔧

By default, tftldr ignores certain resource types:

- Resource types with prefixes: `random_`, `time_`
- Exact resource types: `terraform_data`, `null_resource`

### Configuration Options

You can create a `.tftldr.yml` file in your project directory to specify which resource types to ignore and which columns to display.

```yaml
# .tftldr.yml
ignore:
  # Resource types with these prefixes will be ignored
  prefixes:
    - "random_"
    - "time_"
    - "azurerm_role_"
    
  # These exact resource types will be ignored
  types:
    - "terraform_data"
    - "null_resource"
    - "azurerm_key_vault_secret"

# Column visibility settings
columns:
  changeType: true      # Show the change type (create, update, delete)
  resourceName: true    # Show the resource name
  changedParams: true   # Show parameters that have changed
  resourceType: true    # Show the resource type
  resourceAddress: false # Hide the resource address
```

This allows you to:
- Skip noisy resources like random providers and null resources
- Customize which resource types to ignore based on your needs
- Share configuration across multiple projects


## License 📝
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Disclaimer ⚠️
This tool is designed to help you understand Terraform plans better, but it does not replace the need to review the plan in detail. Always read the Terraform plan output despite what a humours README might suggest.