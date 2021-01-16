package spotify

import "fmt"

func (c *client) getUser() (User, error) {
	usr, err := c.spotify.CurrentUser()
	if err != nil {
		return User{}, fmt.Errorf("Failed to get current user: %w", err)
	}

	return User{
		ID:   usr.ID,
		Name: usr.DisplayName,
	}, nil
}
