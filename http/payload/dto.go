package payload

type UserDto struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Permission int    `json:"permissions"`
}
