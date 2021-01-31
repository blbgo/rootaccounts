package rootaccounts

import (
	"encoding/json"
	"time"

	"github.com/blbgo/record/root"
)

// Account represents an account in the database
type Account interface {
	ID() (uint32, error)
	Details(details *AccountDetails) error
	Update(details *AccountDetails) error
	Delete() error
	WriteNamedValue(name string, value []byte) error
	ReadNamedValue(name string) ([]byte, error)
	RangeNamedValue(cb func(name string, value []byte) bool) error
	DeleteNamedValue(name string) error
}

type account struct {
	root.Item
}

// AccountDetails is a collection of data about an account
type AccountDetails struct {
	AuthLevel  uint32
	Created    time.Time
	LastAccess time.Time
}

func (r account) ID() (uint32, error) {
	return keyToID(r.Item.CopyKey(nil))
}

func (r account) Details(details *AccountDetails) error {
	if details == nil {
		return ErrNilArgument
	}
	return json.Unmarshal(r.Item.Value(), details)
}

func (r account) Update(details *AccountDetails) error {
	value, err := json.Marshal(details)
	if err != nil {
		return err
	}
	return r.Item.UpdateValue(value)
}

func (r account) Delete() error {
	return r.Item.Delete()
}

func (r account) WriteNamedValue(name string, value []byte) error {
	item, err := r.Item.ReadChild([]byte(name))
	if err == nil {
		return item.UpdateValue(value)
	}
	if err == root.ErrItemNotFound {
		return r.Item.QuickChild([]byte(name), value)
	}
	return err
}

func (r account) ReadNamedValue(name string) ([]byte, error) {
	item, err := r.Item.ReadChild([]byte(name))
	if err != nil {
		return nil, err
	}
	return item.Value(), nil
}

func (r account) RangeNamedValue(cb func(name string, value []byte) bool) error {
	var keyBuffer []byte
	return r.Item.RangeChildren(nil, 0, false, func(item root.Item) bool {
		keyBuffer = item.CopyKey(keyBuffer)
		return cb(string(keyBuffer), item.Value())
	})
}

func (r account) DeleteNamedValue(name string) error {
	item, err := r.Item.ReadChild([]byte(name))
	if err != nil {
		return err
	}
	return item.Delete()
}
