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

func (u *user) Name(ctx context.Context) string {
	u.load(ctx)
	return u.NameInternal
}

func (u *user) RealName(ctx context.Context) string {
	u.load(ctx)
	return u.RealNameInternal
}

func (u *user) Email(ctx context.Context) string {
	u.load(ctx)
	return u.EmailInternal
}

func (u *user) load(ctx context.Context) {
	if u.Loaded {
		return
	}

	user, err := u.client.GetUserInfoContext(ctx, u.IDInternal)
	if err != nil {
		// TODO: Hoist up this error somehow
		panic(err)
	}

	u.NameInternal = user.Name
	u.RealNameInternal = user.RealName
	u.EmailInternal = user.Profile.Email
	u.Loaded = true
}
