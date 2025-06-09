# tftldr 🔍

> "Because life's too short to read entire Terraform plans!"

## What is this sorcery? ✨

**tftldr** is your new best friend for making sense of those ridiculously verbose Terraform plan JSON files.

Ever stared at a Terraform plan that's longer than a fantasy novel? Ever wished you could just get the CliffsNotes version? Say no more! **tftldr** transforms that wall of JSON into a beautiful, color-coded table that even your project manager could understand.

## Features 🚀

- Turns intimidating JSON blobs into friendly tables
- Color codes changes (green for creations, yellow for updates, red for deletions)
- Summarizes what's actually changing without the fluff
- Saves your sanity during complex infrastructure deployments
- Makes you look like a Terraform wizard in team meetings

## Installation 📦

```bash
go install github.com/thecomalley/tftldr@latest
```

## Usage 🛠️

Run against your default `tfplan.json` file:

```bash
tftldr
```

Or specify a different plan file:

```bash
tftldr -input path/to/your/plan.json
```

## Development 🧪

To test locally while developing:

```bash
go run ./cmd/tftldr
```

## Why tftldr? 🤔

Because sometimes, you need to know what's changing in your infrastructure without needing to decipher the entire Library of Alexandria worth of JSON. Life's too short for manual diff-ing!

## License 📜

Free as in "free from having to read entire Terraform plans"!

## Contributions 🤝

PRs welcome! Especially if they make the output even more eye-catching or add support for interpretive dance visualizations of your infrastructure changes.
