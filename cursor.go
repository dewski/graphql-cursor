package cursor

import (
	"errors"
	"fmt"

	"github.com/graphql-go/relay"
	"gopkg.in/mgutz/dat.v1"
)

const OrderOnCreatedAt = "created_at"

var (
	DefaultLimit                  = 50
	ErrScopeInvalidBeforeAndAfter = errors.New("You cannot use before and after in the same query")
	ErrScopeInvalidFirstAndLast   = errors.New("You cannot use first and last in the same query")
)

type Cursor interface {
	Cursor() relay.ConnectionCursor
}

type Scope struct {
	relay.ConnectionArguments
	relay.ArraySliceMetaInfo
	Args    map[string]interface{}
	Limit   int
	OrderBy string
	order   string
}

func New() Scope {
	filters := map[string]interface{}{}
	return NewScopeWithFilters(filters)
}

func NewScopeWithFilters(args map[string]interface{}) Scope {
	scope := Scope{
		Args:                args,
		ConnectionArguments: relay.NewConnectionArguments(args),
		Limit:               DefaultLimit,
		OrderBy:             "id",
		order:               "ASC",
	}

	return scope
}

func ApplyScope(builder *dat.SelectBuilder, scope Scope) (*dat.SelectBuilder, error) {
	// Strongly discouraged from using both
	if scope.Before != "" && scope.After != "" {
		return nil, ErrScopeInvalidBeforeAndAfter
	}

	if scope.First != -1 && scope.Last != -1 {
		return nil, ErrScopeInvalidFirstAndLast
	}

	if scope.Before != "" || scope.Last != -1 {
		scope.order = "DESC"
	}

	if scope.After != "" {
		after, err := relay.CursorToOffset(scope.After)
		if err != nil {
			return nil, err
		}

		builder = builder.Where("id > $1", after)
	}

	if scope.First != -1 {
		scope.Limit = scope.First
	}

	if scope.Before != "" {
		before, err := relay.CursorToOffset(scope.Before)
		if err != nil {
			return nil, err
		}

		builder = builder.Where("id < $1", before)
		scope.order = "DESC"
	}

	if scope.Last != -1 {
		scope.Limit = scope.Last
	}

	if scope.Limit != -1 {
		builder = builder.Limit(uint64(scope.Limit + 1))
	}

	builder = ApplyOrder(builder, scope)

	return builder, nil
}

func ApplyOrder(builder *dat.SelectBuilder, scope Scope) *dat.SelectBuilder {
	if scope.OrderBy != "" && scope.order != "" {
		sql := fmt.Sprintf("%s %s", scope.OrderBy, scope.order)
		builder = builder.OrderBy(sql)
	}

	return builder
}

func Connection(arraySlice []Cursor, scope Scope) *relay.Connection {
	if scope.Limit == -1 {
		conn := relay.NewConnection()
		conn.PageInfo = relay.PageInfo{
			StartCursor:     "",
			EndCursor:       "",
			HasPreviousPage: false,
			HasNextPage:     false,
		}

		edges := []*relay.Edge{}
		for _, value := range arraySlice {
			edges = append(edges, &relay.Edge{
				Cursor: value.Cursor(),
				Node:   value,
			})
		}

		conn.Edges = edges

		return conn
	}

	var startCursor, endCursor relay.ConnectionCursor
	args := scope.ConnectionArguments
	// Make sure we're within the bounds of
	limit := min(DefaultLimit, len(arraySlice))
	begin, end := 0, limit

	if args.First != -1 {
		// There are more pages
		if len(arraySlice) == args.First+1 {
			// We don't want to grab the last edge that was used for pagination
			end = end - 1
			endCursor = arraySlice[args.First].Cursor()
		}
	}

	if args.Last != -1 {
		// There are more pages
		if len(arraySlice) == args.Last+1 {
			end = args.Last
			startCursor = arraySlice[args.Last].Cursor()
		}
	}

	slice := arraySlice[begin:end]

	edges := []*relay.Edge{}
	for _, value := range slice {
		edges = append(edges, &relay.Edge{
			Cursor: value.Cursor(),
			Node:   value,
		})
	}

	conn := relay.NewConnection()
	conn.Edges = edges
	conn.PageInfo = relay.PageInfo{
		StartCursor:     startCursor,
		EndCursor:       endCursor,
		HasPreviousPage: startCursor != "",
		HasNextPage:     endCursor != "",
	}

	return conn
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
