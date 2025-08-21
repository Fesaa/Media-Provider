package core

import "github.com/Fesaa/Media-Provider/db/models"

// NotifySubscriptionExhausted adds a notification that the subscription has been fully downloaded
// If this request was a subscription, and one hasn't been sent yet for this download.
// total: a string with the total size (chapters/volumes) of the series
func (c *Core[C, S]) NotifySubscriptionExhausted(total string) {
	if c.Toggles.Toggled(toggleSubscriptionExhausted) || !c.Req.IsSubscription {
		return
	}

	c.Toggles.Toggle(toggleSubscriptionExhausted)
	c.Log.Debug().Msg("Subscription was completed, consider cancelling it")

	body := c.TransLoco.GetTranslation("sub-downloaded-all", c.impl.RefUrl(), c.impl.Title(), total)
	c.Notifier.NotifyContent(c.TransLoco.GetTranslation("sub-downloaded-all-title"), c.impl.Title(), body)
}

// WarnPreferencesFailedToLoad adds a notification if none has been added for this download
func (c *Core[C, S]) WarnPreferencesFailedToLoad() {
	if c.Toggles.Toggled(togglePreferencesFailed) {
		return
	}

	c.Toggles.Toggle(togglePreferencesFailed)
	c.Notifier.NotifyContentQ(
		c.TransLoco.GetTranslation("blacklist-failed-to-load-title", c.impl.Title()),
		c.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
		models.Warning)
}
