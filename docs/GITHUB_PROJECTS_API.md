# GitHub Projects API - Automation Guide

**Goal:** Enable agents to programmatically create tasks on Project boards (with or without GitHub issues)

---

## Prerequisites

```bash
# Add project scope to GitHub CLI
gh auth refresh -s project --hostname github.com

# Verify authentication
gh auth status
```

---

## GitHub Projects v2 API Architecture

GitHub Projects v2 uses **GraphQL API** (not REST). Key concepts:

- **Project** = Board (has ID)
- **ProjectV2Item** = Task/card on board (can be draft or linked to issue)
- **ProjectV2Field** = Custom field (Status, Priority, Size, etc.)
- **ProjectV2FieldValue** = Value for a custom field

---

## Step 1: Get Project IDs

```bash
# List projects (get project ID)
gh api graphql -f query='
query {
  user(login: "sergei-rastrigin") {
    projectsV2(first: 10) {
      nodes {
        id
        number
        title
        url
      }
    }
  }
}'
```

**Expected output:**
```json
{
  "data": {
    "user": {
      "projectsV2": {
        "nodes": [
          {
            "id": "PVT_kwDOABcDE...",
            "number": 1,
            "title": "Ubik Engineering Roadmap",
            "url": "https://github.com/users/sergei-rastrigin/projects/1"
          },
          {
            "id": "PVT_kwDOABcDE...",
            "number": 2,
            "title": "Ubik Business & Marketing",
            "url": "https://github.com/users/sergei-rastrigin/projects/2"
          }
        ]
      }
    }
  }
}
```

**Save these IDs!** You'll need them for all operations.

---

## Step 2: Get Project Field IDs

Each custom field has an ID. You need these to set field values.

```bash
# Get field IDs for a project (replace PROJECT_ID)
gh api graphql -f query='
query {
  node(id: "PROJECT_ID") {
    ... on ProjectV2 {
      fields(first: 20) {
        nodes {
          ... on ProjectV2Field {
            id
            name
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}'
```

**Expected output:**
```json
{
  "fields": {
    "nodes": [
      {
        "id": "PVTF_lADOABcDE...",
        "name": "Status",
        "options": [
          {"id": "PVTO_abc123", "name": "Backlog"},
          {"id": "PVTO_def456", "name": "Ready"},
          {"id": "PVTO_ghi789", "name": "In Progress"}
        ]
      },
      {
        "id": "PVTF_lADOABcDE...",
        "name": "Priority"
      }
    ]
  }
}
```

**Save field IDs and option IDs!**

---

## Step 3: Create Draft Project Item (No Issue)

**Use Case:** Marketing tasks, research tasks, planning items (no code changes)

```bash
# Create draft item on project
gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "PROJECT_ID"
    title: "Campaign: Beta Customer Outreach"
    body: "Acquire 5-10 beta customers through personal network and social media"
  }) {
    projectItem {
      id
    }
  }
}'
```

**Returns:**
```json
{
  "data": {
    "addProjectV2DraftIssue": {
      "projectItem": {
        "id": "PVTI_lADOABcDE..."
      }
    }
  }
}
```

---

## Step 4: Set Custom Field Values

After creating item, set custom fields (Status, Priority, etc.)

```bash
# Set Status field (single select)
gh api graphql -f query='
mutation {
  updateProjectV2ItemFieldValue(input: {
    projectId: "PROJECT_ID"
    itemId: "ITEM_ID"
    fieldId: "STATUS_FIELD_ID"
    value: {
      singleSelectOptionId: "BACKLOG_OPTION_ID"
    }
  }) {
    projectV2Item {
      id
    }
  }
}'

# Set text field (e.g., Dependencies)
gh api graphql -f query='
mutation {
  updateProjectV2ItemFieldValue(input: {
    projectId: "PROJECT_ID"
    itemId: "ITEM_ID"
    fieldId: "DEPENDENCIES_FIELD_ID"
    value: {
      text: "Depends on #1, #2"
    }
  }) {
    projectV2Item {
      id
    }
  }
}'

# Set number field (e.g., Effort)
gh api graphql -f query='
mutation {
  updateProjectV2ItemFieldValue(input: {
    projectId: "PROJECT_ID"
    itemId: "ITEM_ID"
    fieldId: "EFFORT_FIELD_ID"
    value: {
      number: 40
    }
  }) {
    projectV2Item {
      id
    }
  }
}'
```

---

## Step 5: Create Issue and Add to Project

**Use Case:** Engineering tasks (features, bugs) that need code changes

```bash
# 1. Create GitHub issue
ISSUE_ID=$(gh api graphql -f query='
mutation {
  createIssue(input: {
    repositoryId: "REPO_ID"
    title: "Feature: Web UI Dashboard"
    body: "Build agent configuration dashboard"
    labelIds: ["LABEL_ID_1", "LABEL_ID_2"]
    milestoneId: "MILESTONE_ID"
  }) {
    issue {
      id
    }
  }
}' --jq '.data.createIssue.issue.id')

# 2. Add issue to project
gh api graphql -f query='
mutation {
  addProjectV2ItemById(input: {
    projectId: "PROJECT_ID"
    contentId: "'$ISSUE_ID'"
  }) {
    item {
      id
    }
  }
}'
```

---

## Automation Script Template

Create a script for agents to use:

```bash
#!/bin/bash
# create-project-task.sh

PROJECT_ID="$1"
TITLE="$2"
BODY="$3"
STATUS="$4"  # Optional: Backlog, Ready, In Progress, etc.

# Create draft item
ITEM_ID=$(gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "'$PROJECT_ID'"
    title: "'$TITLE'"
    body: "'$BODY'"
  }) {
    projectItem {
      id
    }
  }
}' --jq '.data.addProjectV2DraftIssue.projectItem.id')

echo "Created item: $ITEM_ID"

# Set status if provided
if [ -n "$STATUS" ]; then
  gh api graphql -f query='
  mutation {
    updateProjectV2ItemFieldValue(input: {
      projectId: "'$PROJECT_ID'"
      itemId: "'$ITEM_ID'"
      fieldId: "'$STATUS_FIELD_ID'"
      value: {
        singleSelectOptionId: "'$STATUS_OPTION_ID'"
      }
    }) {
      projectV2Item {
        id
      }
    }
  }'
fi

echo "Task created successfully!"
```

**Usage:**
```bash
./create-project-task.sh \
  "PVT_kwDOABcDE..." \
  "Campaign: Beta Customer Outreach" \
  "Acquire 5-10 beta customers" \
  "Backlog"
```

---

## Integration with AI Agents

### For go-backend-developer Agent

```go
// pkg/github/projects.go
package github

import (
    "bytes"
    "encoding/json"
    "os/exec"
)

type ProjectTask struct {
    ProjectID string
    Title     string
    Body      string
    Status    string
}

func CreateProjectTask(task ProjectTask) (string, error) {
    query := `
    mutation {
      addProjectV2DraftIssue(input: {
        projectId: "%s"
        title: "%s"
        body: "%s"
      }) {
        projectItem {
          id
        }
      }
    }
    `

    cmd := exec.Command("gh", "api", "graphql", "-f",
        "query="+fmt.Sprintf(query, task.ProjectID, task.Title, task.Body))

    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", err
    }

    var result struct {
        Data struct {
            AddProjectV2DraftIssue struct {
                ProjectItem struct {
                    ID string `json:"id"`
                } `json:"projectItem"`
            } `json:"addProjectV2DraftIssue"`
        } `json:"data"`
    }

    json.Unmarshal(out.Bytes(), &result)
    return result.Data.AddProjectV2DraftIssue.ProjectItem.ID, nil
}
```

### For product-strategist Agent

The product-strategist can use the same API to create marketing tasks:

```bash
# Example: Create marketing campaign task
gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "MARKETING_PROJECT_ID"
    title: "Campaign: Beta Customer Outreach"
    body: "Target: 5-10 beta customers by March 1, 2025\n\nTactics:\n- Personal network outreach\n- LinkedIn posts\n- Email warm leads"
  }) {
    projectItem {
      id
    }
  }
}'
```

---

## Configuration File (Store IDs)

Create `.github/project-config.json`:

```json
{
  "projects": {
    "engineering": {
      "id": "PVT_kwDOABcDE...",
      "number": 1,
      "fields": {
        "status": {
          "id": "PVTF_lADOABcDE...",
          "options": {
            "backlog": "PVTO_abc123",
            "ready": "PVTO_def456",
            "in_progress": "PVTO_ghi789",
            "blocked": "PVTO_jkl012",
            "in_review": "PVTO_mno345",
            "done": "PVTO_pqr678"
          }
        },
        "priority": {
          "id": "PVTF_lADOABcDE...",
          "options": {
            "p0": "PVTO_xyz123",
            "p1": "PVTO_xyz456",
            "p2": "PVTO_xyz789",
            "p3": "PVTO_xyz012"
          }
        },
        "effort": {
          "id": "PVTF_lADOABcDE...",
          "type": "number"
        },
        "dependencies": {
          "id": "PVTF_lADOABcDE...",
          "type": "text"
        }
      }
    },
    "marketing": {
      "id": "PVT_kwDOABcDE...",
      "number": 2,
      "fields": {
        "status": {
          "id": "PVTF_lADOABcDE...",
          "options": {
            "backlog": "PVTO_abc789",
            "planning": "PVTO_def012",
            "in_progress": "PVTO_ghi345",
            "in_review": "PVTO_jkl678",
            "launched": "PVTO_mno901",
            "monitoring": "PVTO_pqr234"
          }
        }
      }
    }
  }
}
```

---

## Complete Setup Steps

### 1. Authenticate with project scope
```bash
gh auth refresh -s project --hostname github.com
```

### 2. Get project IDs
```bash
gh api graphql -f query='
query {
  user(login: "sergei-rastrigin") {
    projectsV2(first: 10) {
      nodes {
        id
        number
        title
      }
    }
  }
}' > project-ids.json
```

### 3. Get field IDs for each project
```bash
# For engineering project
gh api graphql -f query='
query {
  node(id: "ENGINEERING_PROJECT_ID") {
    ... on ProjectV2 {
      fields(first: 20) {
        nodes {
          ... on ProjectV2Field {
            id
            name
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}' > engineering-fields.json

# Repeat for marketing project
```

### 4. Create configuration file
```bash
# Parse JSON and create .github/project-config.json
# (Manual step - copy IDs from above outputs)
```

### 5. Test creating a task
```bash
gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "PROJECT_ID"
    title: "Test Task"
    body: "Testing automation"
  }) {
    projectItem {
      id
    }
  }
}'
```

---

## Agent Integration Examples

### go-backend-developer creates engineering task:
```bash
agent-task create \
  --project engineering \
  --title "Feature: Add GET /usage-records endpoint" \
  --priority p1 \
  --size l \
  --labels "area/api,type/feature" \
  --milestone "v0.3.0 - Web UI MVP"
```

### product-strategist creates marketing task:
```bash
agent-task create \
  --project marketing \
  --title "Campaign: Beta Customer Outreach" \
  --status planning \
  --channel email,social \
  --funnel conversion
```

### frontend-developer creates UI task:
```bash
agent-task create \
  --project engineering \
  --title "Component: Agent Configuration Form" \
  --priority p0 \
  --size m \
  --labels "area/web,type/feature" \
  --depends-on "#2"
```

---

## Next Steps

1. **Run setup commands** to get project IDs and field IDs
2. **Create configuration file** with all IDs
3. **Build automation script** (`scripts/create-project-task.sh`)
4. **Test with sample tasks**
5. **Integrate with agent workflows**

---

## Troubleshooting

**Error: "Resource not accessible by integration"**
- Solution: Run `gh auth refresh -s project --hostname github.com`

**Error: "Field not found"**
- Solution: Verify field IDs in project settings

**Error: "Invalid option ID"**
- Solution: Check option IDs for single-select fields

---

## References

- [GitHub Projects V2 API Docs](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [GraphQL API Reference](https://docs.github.com/en/graphql)
- [GitHub CLI Manual](https://cli.github.com/manual/)
