# Integration Tests

This directory contains integration tests for tftldr using real Terraform configurations.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) installed
- Azure CLI logged in (`az login`)
- An active Azure subscription

## Running the Tests

### 1. Set up environment

```bash
cd Integration/azapi
cp .env.example .env
# Edit .env with your Azure subscription ID
```

Or use the Azure CLI to get your subscription ID:
```bash
az account show --query id -o tsv
```

### 2. Initialize and plan

```bash
source .env && export ARM_SUBSCRIPTION_ID
terraform init
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
```

### 3. Run tftldr

From the repository root:
```bash
./tftldr -input Integration/azapi/tfplan.json
```

## Test Cases

### azapi

Tests for Azure API provider resources:

| Resource Type | Description |
|---------------|-------------|
| `azapi_resource` | Azure Container Registry - displays as `Microsoft.ContainerRegistry/registries` |
| `azapi_resource_action` | Start/Stop Azure Spring Apps - displays as `Action: Start` / `Action: Stop` |
| `azurerm_resource_group` | Standard resource group |
| `azurerm_spring_cloud_service` | Azure Spring Apps service |
| `azurerm_user_assigned_identity` | Managed identity |

## Expected Output

```
|--------|---------------------|--------------------|----------------------------------------|----------------------------------------|
| TYPE   | NAME                | CHANGED PARAMETERS | RESOURCE TYPE                          | RESOURCE ADDRESS                       |
|--------|---------------------|--------------------|----------------------------------------|----------------------------------------|
| CREATE | registry1           | All parameters     | Microsoft.ContainerRegistry/registries | azapi_resource.example                 |
| CREATE | Action: Start       | All parameters     | Microsoft.AppPlatform/Spring           | azapi_resource_action.start[0]         |
| CREATE | example-rg          | All parameters     | azurerm_resource_group                 | azurerm_resource_group.example         |
| CREATE | example-springcloud | All parameters     | azurerm_spring_cloud_service           | azurerm_spring_cloud_service.example   |
| CREATE | example             | All parameters     | azurerm_user_assigned_identity         | azurerm_user_assigned_identity.example |
|--------|---------------------|--------------------|----------------------------------------|----------------------------------------|
```

## Notes

- The `.env` file is gitignored to prevent committing secrets
- These tests don't actually deploy resources (plan only)
- `tfplan` and `tfplan.json` are gitignored
