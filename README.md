# grapqhl-cursor

![](https://travis-ci.org/dewski/graphql-cursor.svg?branch=master)

Add GraphQL Relay Cursor Pagination with Postgres.

```go
package web

import (
	"errors"
	"net/http"
	"strconv"

  "github.com/dewski/graphql-cursor"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/labstack/echo"
	"golang.org/x/net/context"
)

var nodeDefinitions *relay.NodeDefinitions

// Each top level type
var tripType *graphql.Object

func (t *Trip) GetEvents(scope cursor.Scope) ([]*TripEvent, error) {
	tripEvents := []*TripEvent{}
	builder := database.Conn().
		Select("*").
		From("trip_events").
		Where("trip_id = $1", t.ID)

	query, err := cursor.ApplyScope(builder, scope)
	if err != nil {
		return nil, err
	}

	err = query.QueryStructs(&tripEvents)
	if err != nil {
		return nil, err
	}

	return tripEvents, nil
}

func main() {
  tripType = graphql.NewObject(graphql.ObjectConfig{
  	Name:        "Trip",
  	Description: "A trip contains the details and event points for a trip",
  	Fields: graphql.Fields{
  		"id": relay.GlobalIDField("Trip", nil),
  		"events": &graphql.Field{
  			Args:        relay.ConnectionArgs,
  			Description: "The events that make up the trip",
  			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
  				scope := cursor.NewScopeWithFilters(p.Args)
  				t := p.Source.(*vehicle.Trip)
  				data := []cursor.Cursor{}

  				events, err := t.GetEvents(scope)
  				if err != nil {
  					return nil, err
  				}

  				for _, event := range events {
  					data = append(data, event)
  				}

  				return cursor.Connection(data, scope), nil
  			},
  		},
  	},
  	Interfaces: []*graphql.Interface{
  		nodeDefinitions.NodeInterface,
  	},
  })

  schema := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
}
```
