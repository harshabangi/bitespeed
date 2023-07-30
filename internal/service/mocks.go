package service

import (
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/stretchr/testify/mock"
	"io"
)

type mockContactStorage struct {
	io.Closer
	mock.Mock
}

func (ms *mockContactStorage) ListContactsByEmailAndPhoneNumber(email string, phoneNumber string) ([]storage.Contact, error) {
	args := ms.Called(email, phoneNumber)
	return args.Get(0).([]storage.Contact), args.Error(1)
}

func (ms *mockContactStorage) ListContactsByID(id int64) ([]storage.Contact, error) {
	args := ms.Called(id)
	return args.Get(0).([]storage.Contact), args.Error(1)
}

func (ms *mockContactStorage) GetContact(id int64) (*storage.Contact, error) {
	args := ms.Called(id)
	return args.Get(0).(*storage.Contact), args.Error(1)
}

func (ms *mockContactStorage) CreateContact(contact storage.Contact) (int64, error) {
	args := ms.Called(contact)
	return args.Get(0).(int64), args.Error(1)
}

func (ms *mockContactStorage) UpdateContact(id int64, contact storage.Contact) error {
	args := ms.Called(id, contact)
	return args.Error(0)
}

func (ms *mockContactStorage) UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(olderContactLinkedID, newerContactLinkedID int64) error {
	args := ms.Called(olderContactLinkedID, newerContactLinkedID)
	return args.Error(0)
}
