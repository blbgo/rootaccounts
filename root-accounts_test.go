package rootaccounts

import (
	"testing"

	"github.com/blbgo/general"
	"github.com/blbgo/record/root"
	"github.com/blbgo/record/store"
	"github.com/blbgo/testing/assert"
)

const testEmail = "test@test.com"
const testPassword = "12345"

func TestOpenCheckNoAccountsClose(t *testing.T) {
	a := assert.New(t)
	store, theRootAccount := setup(a)

	count := 0
	theRootAccount.RangeAccounts(0, false, func(account Account) bool {
		count++
		return true
	})
	a.True(count == 0, "new db has ", count, " accounts??")

	cleanup(a, store)
}

func TestCreateAndReadAccount(t *testing.T) {
	a := assert.New(t)
	store, theRootAccount := setup(a)

	account, err := theRootAccount.CreateAccount(testEmail, testPassword, 5)
	a.NoError(err)
	a.NotNil(account)

	count := 0
	theRootAccount.RangeAccounts(0, false, func(account Account) bool {
		email, err := account.Email()
		a.NoError(err)
		a.True(email == testEmail, "email created as", testEmail, "but is", email)

		details := AccountDetails{}
		err = account.Details(&details)
		a.NoError(err)
		a.True(details.AuthLevel == 5, "auth level was set to 5 but it is", details.AuthLevel)

		err = account.WriteNamedValue("testName", []byte("some text for value"))
		a.NoError(err)

		count++
		return true
	})
	a.True(count == 1, "added 1 account but db has", count, "accounts??")

	account, err = theRootAccount.ReadAccountByEmail(testEmail)
	a.NoError(err)
	a.NotNil(account)

	value, err := account.ReadNamedValue("testName")
	a.NoError(err)
	a.True(string(value) == "some text for value", "value set as \"some text for value\" but is", string(value))

	err = account.Delete()
	a.NoError(err)

	count = 0
	theRootAccount.RangeAccounts(0, false, func(account Account) bool {
		count++
		return true
	})
	a.True(count == 0, "after deleting account db has ", count, " accounts??")

	cleanup(a, store)
}

// helpers
func setup(a *assert.Assert) (store.Store, RootAccount) {
	store, err := store.New(store.NewConfigInMem())
	a.NoError(err)
	a.NotNil(store)

	theRoot := root.New(store)
	a.NotNil(theRoot)

	theRootAccount, err := New(theRoot)
	a.NoError(err)
	a.NotNil(theRootAccount)

	return store, theRootAccount
}

func cleanup(a *assert.Assert, store store.Store) {
	// cleanly close the store
	c, ok := store.(general.DelayCloser)
	a.True(ok)
	doneChan := make(chan error, 100)
	c.Close(doneChan)
	a.Nil(<-doneChan)
}
