package core

// NotifySubscriptionExhausted adds a notification that the subscription has been fully downloaded
// If this request was a subscription, and one hasn't been sent yet for this download.
// total: a string with the total size (chapters/volumes) of the series
func (c *Core[C, S]) NotifySubscriptionExhausted(total string) {
	if c.toggles.Toggled(toggleSubscriptionExhausted) || !c.Req.IsSubscription {
		return
	}

	c.toggles.Toggle(toggleSubscriptionExhausted)
	c.Log.Debug().Msg("Subscription was completed, consider cancelling it")

	body := c.TransLoco.GetTranslation("sub-downloaded-all", c.impl.RefUrl(), c.impl.Title(), total)
	c.Notifier.NotifyContent(c.TransLoco.GetTranslation("sub-downloaded-all-title"), c.impl.Title(), body)
}
