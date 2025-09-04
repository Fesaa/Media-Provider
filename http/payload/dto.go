package payload

import "github.com/Fesaa/Media-Provider/db/models"

type UserDto struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Roles     models.Roles `json:"roles"`
	Pages     []int32      `json:"pages"`
	CanDelete bool         `json:"canDelete"`
}
