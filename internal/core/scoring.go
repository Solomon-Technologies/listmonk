package core

import (
	"encoding/json"
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetScoringRules returns all scoring rules.
func (c *Core) GetScoringRules() (models.ScoringRules, error) {
	var out models.ScoringRules
	if err := c.q.GetScoringRules.Select(&out); err != nil {
		c.log.Printf("error fetching scoring rules: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "scoring rules", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetScoringRule returns a single scoring rule.
func (c *Core) GetScoringRule(id int) (models.ScoringRule, error) {
	var out models.ScoringRule
	if err := c.q.GetScoringRule.Get(&out, id); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "scoring rule", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateScoringRule creates a new scoring rule.
func (c *Core) CreateScoringRule(o models.ScoringRule) (models.ScoringRule, error) {
	if o.Conditions == nil {
		o.Conditions = []byte("{}")
	}

	var id int
	if err := c.q.CreateScoringRule.Get(&id, o.Name, o.Enabled, o.EventType, o.ScoreValue, o.Conditions); err != nil {
		c.log.Printf("error creating scoring rule: %v", err)
		return models.ScoringRule{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "scoring rule", "error", pqErrMsg(err)))
	}

	return c.GetScoringRule(id)
}

// UpdateScoringRule updates a scoring rule.
func (c *Core) UpdateScoringRule(id int, o models.ScoringRule) (models.ScoringRule, error) {
	if o.Conditions == nil {
		o.Conditions = []byte("{}")
	}

	res, err := c.q.UpdateScoringRule.Exec(id, o.Name, o.Enabled, o.EventType, o.ScoreValue, o.Conditions)
	if err != nil {
		c.log.Printf("error updating scoring rule: %v", err)
		return models.ScoringRule{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "scoring rule", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.ScoringRule{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "scoring rule"))
	}

	return c.GetScoringRule(id)
}

// DeleteScoringRule deletes a scoring rule.
func (c *Core) DeleteScoringRule(id int) error {
	res, err := c.q.DeleteScoringRule.Exec(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "scoring rule", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "scoring rule"))
	}
	return nil
}

// ApplyScore applies a score change to a subscriber based on an event.
func (c *Core) ApplyScore(subscriberID int, eventType string, meta map[string]any) error {
	// Get matching rules.
	var rules models.ScoringRules
	if err := c.q.GetScoringRulesByEvent.Select(&rules, eventType); err != nil {
		c.log.Printf("error fetching scoring rules for event %s: %v", eventType, err)
		return err
	}

	for _, rule := range rules {
		// Apply the score change.
		var newScore int
		if err := c.q.UpdateSubscriberScore.Get(&newScore, subscriberID, rule.ScoreValue); err != nil {
			c.log.Printf("error updating subscriber score: %v", err)
			continue
		}

		// Log the change.
		metaJSON, _ := json.Marshal(meta)
		c.q.InsertScoreLog.Exec(subscriberID, rule.ID, eventType, rule.ScoreValue, newScore, metaJSON)
	}

	return nil
}

// GetSubscriberScoreLog returns the score history for a subscriber.
func (c *Core) GetSubscriberScoreLog(subscriberID, offset, limit int) (models.ScoreLogs, error) {
	var out models.ScoreLogs
	if err := c.db.Select(&out, c.q.GetSubscriberScoreLog, subscriberID, offset, limit); err != nil {
		c.log.Printf("error fetching score log: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "score log", "error", pqErrMsg(err)))
	}
	return out, nil
}

// DecayInactiveScores reduces scores for inactive subscribers.
func (c *Core) DecayInactiveScores(decayAmount int) (int, error) {
	rows, err := c.db.Exec(c.q.DecayInactiveScores, decayAmount)
	if err != nil {
		c.log.Printf("error decaying scores: %v", err)
		return 0, err
	}
	n, _ := rows.RowsAffected()
	return int(n), nil
}
