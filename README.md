# Aigile

Aigile is a CLI tool that helps you generate User Stories from an XLSX file using LLM and create them in GitHub/Azure DevOps.

## Installation

```bash
go install github.com/leocomelli/aigile@latest
```

## Usage

### GitHub

1. Create a GitHub token with the following scopes:
   - `repo` (Full control of private repositories)
   - `project` (Full control of projects)
   - `read:org` (Read organization data)

   > **Important Note**: Currently, when using a personal GitHub account, the Fine-Grained Token interface does not allow setting the Projects scope. As a workaround, you need to use a Classic Personal Access Token (PAT) instead. This is a known limitation reported by many users in the GitHub Community Discussions.

2. Set the token in your environment:
   ```bash
   export GITHUB_TOKEN=your_token
   ```

3. Run the command:
   ```bash
   aigile generate --provider github --owner your_username --repo your_repo --file path/to/your/file.xlsx
   ```

### Azure DevOps

1. Create a Personal Access Token (PAT) with the following scopes:
   - `Work Items (Read & Write)`
   - `Project and Team (Read)`

2. Set the token in your environment:
   ```bash
   export AZURE_DEVOPS_TOKEN=your_token
   export AZURE_DEVOPS_ORG=your_organization
   export AZURE_DEVOPS_PROJECT=your_project
   ```

3. Run the command:
   ```bash
   aigile generate --provider azure --file path/to/your/file.xlsx
   ```

## XLSX File Format

The XLSX file should have the following columns:

- `Title`: The title of the User Story
- `Description`: The description of the User Story
- `Acceptance Criteria`: The acceptance criteria of the User Story
- `Project`: The name of the project to add the User Story to (optional)
- `Parent Feature`: The ID of the parent feature (optional)

## Features

- Generate User Stories from an XLSX file using LLM
- Create User Stories in GitHub/Azure DevOps
- Add User Stories to projects
- Link User Stories to parent features
- Generate suggested tasks for User Stories (optional)

## License

MIT 