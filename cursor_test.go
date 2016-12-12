package cursor

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"github.com/graphql-go/relay"
	"github.com/stretchr/testify/assert"
)

func TestScopeWithFiltersSetsDefaults(t *testing.T) {
	args := map[string]interface{}{}

	scope := NewScopeWithFilters(args)
	assert.Equal(t, DefaultLimit, scope.Limit)
	assert.Equal(t, "id", scope.OrderBy)
}

type graphqlRecord struct {
	ID int64 `db:"id"`
}

func (gr *graphqlRecord) Cursor() relay.ConnectionCursor {
	str := fmt.Sprintf("%v%v", relay.PREFIX, gr.ID)
	return relay.ConnectionCursor(base64.StdEncoding.EncodeToString([]byte(str)))
}

func (gr *graphqlRecord) ToGlobalID() string {
	id := strconv.FormatInt(gr.ID, 10)
	return relay.ToGlobalID("graphqlRecord", id)
}

func GetRecords(scope Scope) ([]*graphqlRecord, error) {
	var records []*graphqlRecord
	query := Conn().
		Select("*").
		From("graphql_records")

	query, err := ApplyScope(query, scope)
	if err != nil {
		return nil, err
	}

	err = query.QueryStructs(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func TestBeforeAndAfterIncludedInArgments(t *testing.T) {
	// Test goes here
	args := map[string]interface{}{
		"before": "12234",
		"after":  "12344",
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)
	assert.Nil(t, results)
	assert.Equal(t, ErrScopeInvalidBeforeAndAfter, err)
}

func TestFirstAndLastIncludedInArgments(t *testing.T) {
	// Test goes here
	args := map[string]interface{}{
		"first": 10,
		"last":  10,
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)
	assert.Nil(t, results)
	assert.Equal(t, ErrScopeInvalidFirstAndLast, err)
}

func TestRecordsWithFilter_AfterFirst(t *testing.T) {
	// Test goes here
	three := graphqlRecord{ID: 3}
	args := map[string]interface{}{
		"after": three.Cursor(),
		"first": 3,
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 4},
		&graphqlRecord{ID: 5},
		&graphqlRecord{ID: 6},
		&graphqlRecord{ID: 7},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestRecordsWithFilter_After(t *testing.T) {
	// Test goes here
	record := graphqlRecord{ID: 15}
	args := map[string]interface{}{
		"after": record.Cursor(),
	}

	scope := NewScopeWithFilters(args)
	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 16},
		&graphqlRecord{ID: 17},
		&graphqlRecord{ID: 18},
		&graphqlRecord{ID: 19},
		&graphqlRecord{ID: 20},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestRecordsWithFilter_First(t *testing.T) {
	// Test goes here
	args := map[string]interface{}{
		"first": 5,
	}

	scope := NewScopeWithFilters(args)
	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 1},
		&graphqlRecord{ID: 2},
		&graphqlRecord{ID: 3},
		&graphqlRecord{ID: 4},
		&graphqlRecord{ID: 5},
		&graphqlRecord{ID: 6},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestRecordsWithFilter_BeforeLast(t *testing.T) {
	// Test goes here
	record := graphqlRecord{ID: 20}
	args := map[string]interface{}{
		"before": record.Cursor(),
		"last":   3,
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 19},
		&graphqlRecord{ID: 18},
		&graphqlRecord{ID: 17},
		&graphqlRecord{ID: 16},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestRecordsWithFilter_Before(t *testing.T) {
	// Test goes here
	record := graphqlRecord{ID: 5}
	args := map[string]interface{}{
		"before": record.Cursor(),
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 4},
		&graphqlRecord{ID: 3},
		&graphqlRecord{ID: 2},
		&graphqlRecord{ID: 1},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestRecordsWithFilter_Last(t *testing.T) {
	// Test goes here
	args := map[string]interface{}{
		"last": 5,
	}

	scope := NewScopeWithFilters(args)

	results, err := GetRecords(scope)

	expected := []*graphqlRecord{
		&graphqlRecord{ID: 20},
		&graphqlRecord{ID: 19},
		&graphqlRecord{ID: 18},
		&graphqlRecord{ID: 17},
		&graphqlRecord{ID: 16},
		&graphqlRecord{ID: 15},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}
