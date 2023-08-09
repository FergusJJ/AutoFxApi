package storage_test

import (
	"api/internal/storage"
	"fmt"
	"testing"
)

func Test_postgres(t *testing.T) {
	store, err := storage.NewPostgresStore()
	if err != nil {
		t.Fatalf("storage.NewPostgresStore(): %+v", err)
	}
	fmt.Printf("%+v\n", store)
}
