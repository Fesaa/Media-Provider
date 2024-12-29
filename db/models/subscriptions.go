package models

type Subscriptions interface {
	All() ([]Subscription, error)
	Get(uint) (*Subscription, error)

	New(Subscription) (*Subscription, error)
	Update(Subscription) error
	Delete(uint) error
}
