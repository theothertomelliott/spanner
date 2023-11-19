package slack

import (
	"context"
)

type user struct {
	client socketClient
	Loaded bool `json:"loaded"`

	IDInternal       string `json:"id"`
	NameInternal     string `json:"display_name"`
	RealNameInternal string `json:"real_name"`
	EmailInternal    string `json:"email"`
}

func (u *user) ID() string {
	return u.IDInternal
}

func (u *user) Name() string {
	u.load()
	return u.NameInternal
}

func (u *user) RealName() string {
	u.load()
	return u.RealNameInternal
}

func (u *user) Email() string {
	u.load()
	return u.EmailInternal
}

func (u *user) load() {
	if u.Loaded {
		return
	}

	user, err := u.client.GetUserInfoContext(context.TODO(), u.IDInternal)
	if err != nil {
		// TODO: Hoist up this error somehow
		panic(err)
	}

	u.NameInternal = user.Name
	u.RealNameInternal = user.RealName
	u.EmailInternal = user.Profile.Email
	u.Loaded = true
}
