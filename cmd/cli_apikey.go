package main

import (
	"fmt"
	"os"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/knadh/listmonk/internal/core"
	"gopkg.in/volatiletech/null.v6"
)

// runCreateAPIUser handles the --create-api-user CLI flag. It mints a new
// tenant-scoped API user via core.CreateUser (which generates the 32-byte
// random token internally), prints the username + one-time token to stdout,
// and exits.
//
// Required flags: --username, --company-id, --role-id.
//
// Why a CLI subcommand at all when the UI already creates API users?
// 1. Deterministic, scriptable provisioning during fresh tenant setup.
// 2. Recovery path when the only existing API key is lost or revoked and the
//    UI is locked out (no working session).
// 3. Useful in the multi-tenant rollout window where the UI hasn't caught up
//    with the company-picker affordance yet.
func runCreateAPIUser(c *core.Core) {
	username := ko.String("username")
	companyID := ko.Int("company-id")
	roleID := ko.Int("role-id")
	name := ko.String("name")

	if username == "" {
		fmt.Fprintln(os.Stderr, "error: --username is required")
		os.Exit(1)
	}
	if companyID <= 0 {
		fmt.Fprintln(os.Stderr, "error: --company-id is required and must be > 0")
		os.Exit(1)
	}
	if roleID <= 0 {
		fmt.Fprintln(os.Stderr, "error: --role-id is required and must be > 0")
		fmt.Fprintln(os.Stderr, "  hint: query roles with `psql ... -c \"SELECT id, name, company_id FROM roles WHERE company_id = <id>;\"`")
		os.Exit(1)
	}

	// Validate the role belongs to the same company. Cross-tenant role
	// assignment would silently grant access into a tenant the API user
	// doesn't belong to.
	role, err := c.GetRole(roleID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: role-id=%d not found: %v\n", roleID, err)
		os.Exit(1)
	}
	if role.CompanyID != companyID {
		fmt.Fprintf(os.Stderr, "error: role-id=%d belongs to company_id=%d, not the requested company_id=%d\n", roleID, role.CompanyID, companyID)
		os.Exit(1)
	}

	if name == "" {
		name = username
	}

	u := auth.User{
		Username:   username,
		Name:       name,
		Type:       auth.UserTypeAPI,
		Status:     auth.UserStatusEnabled,
		UserRoleID: roleID,
		CompanyID:  companyID,
		// Email + Password are set by core.CreateUser for API users.
		Email:    null.String{},
		Password: null.String{},
	}

	created, err := c.CreateUser(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: create-api-user failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ API user created.")
	fmt.Printf("  username:   %s\n", created.Username)
	fmt.Printf("  company_id: %d\n", created.CompanyID)
	fmt.Printf("  role_id:    %d\n", created.UserRoleID)
	fmt.Printf("  user_id:    %d\n", created.ID)
	fmt.Println()
	fmt.Println("  TOKEN (save this NOW — it will not be shown again):")
	fmt.Printf("    %s\n", created.Password.String)
	fmt.Println()
	fmt.Println("  Use as HTTP Basic Auth:")
	fmt.Printf("    curl -u '%s:%s' https://your.listmonk.host/api/lists\n", created.Username, created.Password.String)
	fmt.Println()
	fmt.Println("  ⚠ The running listmonk service has cached the prior user set.")
	fmt.Println("    Restart it so the new key authenticates:")
	fmt.Println("      systemctl restart listmonk    # or pm2 restart listmonk")
	fmt.Println()

	os.Exit(0)
}
