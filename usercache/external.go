package usercache

import (
	"github.com/pkg/errors"
	"github.com/tychoish/emt"
	"github.com/tychoish/gimlet"
)

// ExternalOptions provides functions to inject the functionality of the user
// cache from an external source.
type ExternalOptions struct {
	PutUserGetToken PutUserGetToken
	GetUserByToken  GetUserByToken
	ClearUserToken  ClearUserToken
	GetUserByID     GetUserByID
	GetOrCreateUser GetOrCreateUser
}

func (opts ExternalOptions) Validate() error {
	catcher := emt.NewBasicCatcher()
	catcher.NewWhen(opts.PutUserGetToken == nil, "PutUserGetToken must be defined")
	catcher.NewWhen(opts.GetUserByToken == nil, "GetUserByToken must be defined")
	catcher.NewWhen(opts.ClearUserToken == nil, "ClearUserToken must be defined")
	catcher.NewWhen(opts.GetUserByID == nil, "GetUserByID must be defined")
	catcher.NewWhen(opts.GetOrCreateUser == nil, "GetOrCreateUser must be defined")
	return catcher.Resolve()
}

// NewExternal returns an external user cache.
func NewExternal(opts ExternalOptions) (Cache, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid cache options")
	}
	return &ExternalCache{Opts: opts}, nil
}

type ExternalCache struct {
	Opts ExternalOptions
}

func (c *ExternalCache) Add(u gimlet.User) error           { _, err := c.Opts.GetOrCreateUser(u); return err }
func (c *ExternalCache) Put(u gimlet.User) (string, error) { return c.Opts.PutUserGetToken(u) }
func (c *ExternalCache) Get(token string) (gimlet.User, bool, error) {
	return c.Opts.GetUserByToken(token)
}
func (c *ExternalCache) Clear(u gimlet.User, all bool) error       { return c.Opts.ClearUserToken(u, all) }
func (c *ExternalCache) Find(id string) (gimlet.User, bool, error) { return c.Opts.GetUserByID(id) }
func (c *ExternalCache) GetOrCreate(u gimlet.User) (gimlet.User, error) {
	return c.Opts.GetOrCreateUser(u)
}
