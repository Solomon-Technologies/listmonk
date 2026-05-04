package models

import null "gopkg.in/volatiletech/null.v6"

// Company is a tenant in the multi-tenant fork (v7.17.0+).
type Company struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt null.Time `db:"created_at" json:"created_at"`
	UpdatedAt null.Time `db:"updated_at" json:"updated_at"`
}

// CompanyStats is a Company plus row counts of dependent tenant data.
// Returned by GET /api/companies/stats so the admin Companies page can show
// "this tenant has 5 lists / 6,915 subscribers / ..." before a destructive
// delete.
type CompanyStats struct {
	ID               int    `db:"id" json:"id"`
	Name             string `db:"name" json:"name"`
	Slug             string `db:"slug" json:"slug"`
	Users            int    `db:"users" json:"users"`
	Lists            int    `db:"lists" json:"lists"`
	Subscribers      int    `db:"subscribers" json:"subscribers"`
	Campaigns        int    `db:"campaigns" json:"campaigns"`
	Templates        int    `db:"templates" json:"templates"`
	WarmingSenders   int    `db:"warming_senders" json:"warming_senders"`
	WarmingCampaigns int    `db:"warming_campaigns" json:"warming_campaigns"`
	Roles            int    `db:"roles" json:"roles"`
}
