package rootaccounts

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/blbgo/record/root"
)

// RootAccount provides access to a database of accounts
type RootAccount interface {
	CreateAccount(email string, authLevel uint32) (Account, error)
	ReadAccount(id uint32) (Account, error)
	ReadAccountByEmail(email string) (Account, error)
	RangeAccounts(startID uint32, reverse bool, cb func(account Account) bool) error
}

// ErrInvalidIDInDatabase bar record value is wrong length
var ErrInvalidIDInDatabase = errors.New("Invalid account ID in database")

// ErrAlreadyExists3Attempts Attempted to create new account 3 times
var ErrAlreadyExists3Attempts = errors.New("Attempted to create new account 3 times")

// ErrNilArgument Argument is nil
var ErrNilArgument = errors.New("Argument is nil")

type rootAccount struct {
	root.Item
}

// New creates a RootAccount implemented by rootAccount
func New(theRoot root.Root) (RootAccount, error) {
	item, err := theRoot.RootItem(
		"github.com/blbgo/rootaccount",
		"github.com/blbgo/rootaccount root item",
	)
	if err != nil {
		return nil, err
	}
	return rootAccount{Item: item}, nil
}

// *** implement RootAccount

var maxID = []byte{0xff, 0xff, 0xff, 0xff}

func (r rootAccount) CreateAccount(email string, authLevel uint32) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	for try := 0; try < 3; try++ {
		newAccount, err := r.tryCreateAccount(email, authLevel)
		if err == nil {
			return newAccount, nil
		}
		if err != root.ErrAlreadyExists {
			return nil, err
		}
	}
	return nil, ErrAlreadyExists3Attempts
}

func (r rootAccount) ReadAccount(id uint32) (Account, error) {
	var key [4]byte
	binary.BigEndian.PutUint32(key[:], id)
	item, err := r.ReadChild(key[:])
	if err != nil {
		return nil, err
	}
	return account{Item: item}, nil
}

func (r rootAccount) ReadAccountByEmail(email string) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	item, err := r.Item.ReadChildByIndex([]byte(email))
	if err != nil {
		return nil, err
	}
	return account{Item: item}, nil
}

func (r rootAccount) RangeAccounts(
	startID uint32,
	reverse bool,
	cb func(account Account) bool,
) error {
	var startKey [4]byte
	binary.BigEndian.PutUint32(startKey[:], startID)
	return r.RangeChildren(startKey[:], 0, reverse, func(item root.Item) bool {
		return cb(account{Item: item})
	})
}

// *** helpers

func (r rootAccount) tryCreateAccount(email string, authLevel uint32) (Account, error) {
	key := make([]byte, 4)
	err := r.RangeChildren(maxID, 0, true, func(item root.Item) bool {
		key = item.CopyKey(key)
		return false
	})
	if err != nil {
		return nil, err
	}
	id, err := keyToID(key)
	if err != nil {
		return nil, err
	}
	id++
	binary.BigEndian.PutUint32(key, id)
	now := time.Now()
	details := &AccountDetails{AuthLevel: authLevel, Created: now, LastAccess: now}
	value, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	item, err := r.Item.CreateChild(key, value, [][]byte{[]byte(email)})
	if err != nil {
		return nil, err
	}
	return account{Item: item}, nil
}

func keyToID(key []byte) (uint32, error) {
	if len(key) != 4 {
		return 0, ErrInvalidIDInDatabase
	}
	return binary.BigEndian.Uint32(key), nil
}
