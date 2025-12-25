package skills

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewSkillsCommand creates the skills command group with dependencies from the container.
func NewSkillsCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage Claude Code skills",
		Long:  "View available Claude Code skills and manage skill access.",
	}

	cmd.AddCommand(NewListCommand(c))
	cmd.AddCommand(NewShowCommand(c))
	cmd.AddCommand(NewMyCommand(c))

	return cmd
}

// NewListCommand creates the skills list command with dependencies from the container.
func NewListCommand(c *container.Container) *cobra.Command {
	var showLocal bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available skills",
		Long:  "Display all available skills from the platform catalog or locally installed skills.",
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsService, err := c.SkillsService()
			if err != nil {
				return fmt.Errorf("failed to get skills service: %w", err)
			}

			// If showing local skills, no need to authenticate
			if showLocal {
				localSkills, err := skillsService.GetLocalSkills()
				if err != nil {
					return fmt.Errorf("failed to get local skills: %w", err)
				}

				if len(localSkills) == 0 {
					fmt.Println("No local skills found in .claude/skills/")
					fmt.Println("\nRun 'arfa sync' to fetch skills from the platform.")
					return nil
				}

				fmt.Printf("\nInstalled Skills (%d):\n\n", len(localSkills))

				// Create table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tVERSION\tFILES\tCATEGORY")
				fmt.Fprintln(w, "────\t───────\t─────\t────────")

				for _, skill := range localSkills {
					version := skill.Version
					if version == "" {
						version = "unknown"
					}
					category := skill.Category
					if category == "" {
						category = "-"
					}
					fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", skill.Name, version, len(skill.Files), category)
				}

				w.Flush()
				fmt.Println()
				fmt.Println("💡 Tip: Use 'arfa skills show <name>' to see details for a specific skill")
				fmt.Println()

				return nil
			}

			// For catalog skills, require authentication
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			ctx := context.Background()
			skills, err := skillsService.ListCatalogSkills(ctx)
			if err != nil {
				return fmt.Errorf("failed to list skills: %w", err)
			}

			if len(skills) == 0 {
				fmt.Println("No skills available in the platform catalog.")
				return nil
			}

			fmt.Printf("\nAvailable Skills (%d):\n\n", len(skills))

			// Create table writer
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tCATEGORY\tDESCRIPTION")
			fmt.Fprintln(w, "────\t───────\t────────\t───────────")

			for _, skill := range skills {
				description := skill.Description
				if len(description) > 60 {
					description = description[:57] + "..."
				}
				category := skill.Category
				if category == "" {
					category = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", skill.Name, skill.Version, category, description)
			}

			w.Flush()
			fmt.Println()
			fmt.Println("💡 Tip: Use 'arfa skills show <name>' to see skill details")
			fmt.Println("        Use 'arfa skills my' to see your assigned skills")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&showLocal, "local", false, "Show locally installed skills only")

	return cmd
}

// NewShowCommand creates the skills show command with dependencies from the container.
func NewShowCommand(c *container.Container) *cobra.Command {
	var showLocal bool

	cmd := &cobra.Command{
		Use:   "show <skill-name>",
		Short: "Show skill details",
		Long:  "Display detailed information about a specific skill.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			skillsService, err := c.SkillsService()
			if err != nil {
				return fmt.Errorf("failed to get skills service: %w", err)
			}

			// If showing local skill, no authentication needed
			if showLocal {
				skill, err := skillsService.GetLocalSkill(skillName)
				if err != nil {
					return fmt.Errorf("failed to get local skill: %w", err)
				}

				fmt.Printf("\nSkill: %s\n", skill.Name)
				if skill.Version != "" {
					fmt.Printf("Version: %s\n", skill.Version)
				}
				if skill.Description != "" {
					fmt.Printf("Description: %s\n", skill.Description)
				}
				if skill.Category != "" {
					fmt.Printf("Category: %s\n", skill.Category)
				}

				fmt.Printf("\nFiles (%d):\n", len(skill.Files))
				for _, file := range skill.Files {
					fmt.Printf("  - %s\n", file)
				}

				status := "✓ Installed"
				if !skill.IsEnabled {
					status = "✗ Disabled"
				}
				fmt.Printf("\nStatus: %s\n", status)
				fmt.Printf("Location: .claude/skills/%s/\n", skill.Name)
				if skill.InstalledAt != "" {
					fmt.Printf("Installed: %s\n", skill.InstalledAt)
				}
				fmt.Println()

				return nil
			}

			// For catalog/employee skills, require authentication
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			// Try to find skill in catalog by name
			ctx := context.Background()
			skill, err := skillsService.GetSkillByName(ctx, skillName)
			if err != nil {
				return fmt.Errorf("failed to get skill: %w", err)
			}

			fmt.Printf("\nSkill: %s\n", skill.Name)
			fmt.Printf("Version: %s\n", skill.Version)
			fmt.Printf("Description: %s\n", skill.Description)
			fmt.Printf("Category: %s\n", skill.Category)

			if len(skill.Files) > 0 {
				fmt.Printf("\nFiles (%d):\n", len(skill.Files))
				for _, file := range skill.Files {
					fmt.Printf("  - %s\n", file.Path)
				}
			}

			// Display dependencies if any
			if skill.Dependencies != nil && len(skill.Dependencies) > 0 {
				fmt.Println("\nDependencies:")
				if mcpServers, ok := skill.Dependencies["mcp_servers"].([]interface{}); ok && len(mcpServers) > 0 {
					fmt.Println("  MCP Servers:")
					for _, server := range mcpServers {
						fmt.Printf("    - %v\n", server)
					}
				}
				if skills, ok := skill.Dependencies["skills"].([]interface{}); ok && len(skills) > 0 {
					fmt.Println("  Skills:")
					for _, s := range skills {
						fmt.Printf("    - %v\n", s)
					}
				}
			}

			status := "Available"
			if !skill.IsActive {
				status = "Inactive"
			}
			fmt.Printf("\nStatus: %s\n", status)
			fmt.Printf("ID: %s\n", skill.ID)
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&showLocal, "local", false, "Show locally installed skill details")

	return cmd
}

// NewMyCommand creates the skills my command with dependencies from the container.
func NewMyCommand(c *container.Container) *cobra.Command {
	var showDetails bool

	cmd := &cobra.Command{
		Use:   "my",
		Short: "List your assigned skills",
		Long:  "Display skills that have been assigned to you by your organization.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Require authentication
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			skillsService, err := c.SkillsService()
			if err != nil {
				return fmt.Errorf("failed to get skills service: %w", err)
			}

			ctx := context.Background()
			skills, err := skillsService.ListEmployeeSkills(ctx)
			if err != nil {
				return fmt.Errorf("failed to list employee skills: %w", err)
			}

			if len(skills) == 0 {
				fmt.Println("\nNo skills assigned to you yet.")
				fmt.Println("\nContact your administrator to request skills.")
				return nil
			}

			fmt.Printf("\nYour Assigned Skills (%d):\n\n", len(skills))

			if showDetails {
				// Show detailed view
				for i, skill := range skills {
					if i > 0 {
						fmt.Println(strings.Repeat("─", 60))
					}

					fmt.Printf("Name: %s\n", skill.Name)
					fmt.Printf("Version: %s\n", skill.Version)
					fmt.Printf("Description: %s\n", skill.Description)
					fmt.Printf("Category: %s\n", skill.Category)

					status := "✓ Enabled"
					if !skill.IsEnabled {
						status = "✗ Disabled"
					}
					fmt.Printf("Status: %s\n", status)

					if len(skill.Files) > 0 {
						fmt.Printf("Files: %d\n", len(skill.Files))
					}

					if skill.InstalledAt != nil {
						fmt.Printf("Installed: %s\n", skill.InstalledAt.Format("2006-01-02 15:04:05"))
					}
					fmt.Println()
				}
			} else {
				// Show table view
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tCATEGORY")
				fmt.Fprintln(w, "────\t───────\t──────\t────────")

				for _, skill := range skills {
					status := "✓ enabled"
					if !skill.IsEnabled {
						status = "✗ disabled"
					}
					category := skill.Category
					if category == "" {
						category = "-"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", skill.Name, skill.Version, status, category)
				}

				w.Flush()
				fmt.Println()
				fmt.Println("💡 Tip: Use 'arfa skills show <name>' to see skill details")
				fmt.Println("        Use 'arfa sync' to install assigned skills locally")
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed information for each skill")

	return cmd
}
