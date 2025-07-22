package payload

type UserDto struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Permission int    `json:"permissions"`
	CanDelete  bool   `json:"canDelete"`
}
