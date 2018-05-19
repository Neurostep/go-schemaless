package storagetest

import (
	"context"
	"errors"

	"github.com/rbastic/go-schemaless"
	"github.com/rbastic/go-schemaless/models"
	"testing"
	"time"
)

const (
	sqlDateFormat = "2006-01-02 15:04:05" // TODO: Hmm, should we make this a constant somewhere? Likely.

	cellID = "hello-001"
	baseCol = "BASE"
	otherCellID = "hello"
	testString = "The shaved yak drank from the bitter well"
	testString2 = "The printer is on fire"
	testString3 = "The appropriate printer-fire-response-team has been notified"
)

type Errstore struct{}

func (e Errstore) Get(key string) ([]byte, bool, error) {
	return nil, false, errors.New("error storage get")
}
func (e Errstore) Set(key string, val []byte) error { return errors.New("error storage Set") }
func (e Errstore) Delete(key string) (bool, error)  { return false, errors.New("error storage Delete") }
func (e Errstore) ResetConnection(key string) error {
	return errors.New("error storage ResetConnection")
}

func runPuts( t * testing.T, storage schemaless.Storage) {
	err := storage.PutCell(context.TODO(), cellID, baseCol, 1, models.Cell{Body: []byte(testString)})
	if err != nil {
		t.Fatal(err)
	}

	err = storage.PutCell(context.TODO(), cellID, baseCol, 2, models.Cell{Body: []byte(testString2)})
	if err != nil {
		t.Fatal(err)
	}

	err = storage.PutCell(context.TODO(), cellID, baseCol, 3, models.Cell{Body: []byte(testString3)})
	if err != nil {
		t.Fatal(err)
	}
}

// StorageTest is a simple sanity check for a schemaless Storage backend
func StorageTest(t *testing.T, storage schemaless.Storage) {
	startTime := time.Now().Format(sqlDateFormat)

	time.Sleep(time.Second * 1)

	defer storage.Destroy(context.TODO())
	v, ok, err := storage.GetCell(context.TODO(), otherCellID, baseCol, 1)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Errorf("getting a non-existent key was 'ok': v=%v ok=%v\n", v, ok)
	}

	runPuts(t, storage)

	v, ok, err = storage.GetCellLatest(context.TODO(), cellID, baseCol)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || string(v.Body) != testString3 {
		t.Errorf("failed getting a valid key: v=%v ok=%v\n", v, ok)
	}

	v, ok, err = storage.GetCell(context.TODO(), cellID, baseCol, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || string(v.Body) != testString {
		t.Errorf("GetCell failed when retrieving an old value: v=%v ok=%v\n", v, ok)
	}

	var cells []models.Cell
	cells, ok, err = storage.GetCellsForShard(context.TODO(), 0, "timestamp", startTime, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected a slice of cells, response was:", cells)
	}

	if len(cells) == 0 {
		t.Fatal("we have an obvious problem")
	}

	err = storage.ResetConnection(context.TODO(), otherCellID)
	if err != nil {
		t.Errorf("failed resetting connection for key: err=%v\n", err)
	}
}
