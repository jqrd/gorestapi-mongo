package mainrpc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/jqrd/gorestapi-mongo/mocks"
	"github.com/jqrd/gorestapi-mongo/model/db"
	"github.com/jqrd/gorestapi-mongo/model/svc"
)

func TestThingPost(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	store := new(mocks.DataStore)
	err := Setup(r, store)
	assert.Nil(t, err)

	// Create Item
	widgets := make([]*svc.ThingWidget, 0)
	input := &svc.Thing{
		Name:        "name",
		Description: "descr",
		Widgets:     widgets,
	}

	bytes, _ := bson.Marshal(input)
	expected := &svc.Thing{}
	bson.Unmarshal(bytes, expected)
	expected.Id = "a42"

	// Mock calls to data store
	thingsCol := new(mocks.MongoCollection[*db.Thing])
	store.On("Things").Return(thingsCol)
	thingsCol.On("InsertOne", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { args.Get(1).(*db.Thing).Id = expected.Id }).
		Return(nil)

	// Make request and validate we get back proper response
	e := httpexpect.Default(t, server.URL)
	e.POST("/api/things").WithJSON(input).Expect().
		Status(http.StatusOK).
		JSON().Object().Equal(expected)

	// Check remaining expectations
	store.AssertExpectations(t)
	thingsCol.AssertExpectations(t)

}

// TODO also test post thing with widgets

/* TODO
func TestThingsFind(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	grs := new(mocks.GRStore)
	err := Setup(r, grs)
	assert.Nil(t, err)

	// Return Item
	i := []*gorestapi.Thing{
		{
			ID:   "id1",
			Name: "name1",
		},
		{
			ID:   "id2",
			Name: "name2",
		},
	}

	// Mock call to item store
	grs.On("ThingsFind", mock.Anything, mock.AnythingOfType("*queryp.QueryParameters")).Once().Return(i, int64(2), nil)

	// Make request and validate we get back proper response
	e := httpexpect.New(t, server.URL)
	e.GET("/api/things").Expect().Status(http.StatusOK).JSON().Object().Equal(&store.Results{Count: 2, Results: i})

	// Check remaining expectations
	grs.AssertExpectations(t)

}

func TestThingGetByID(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	grs := new(mocks.GRStore)
	err := Setup(r, grs)
	assert.Nil(t, err)

	// Create Item
	i := &gorestapi.Thing{
		ID:   "id",
		Name: "name",
	}

	// Mock call to item store
	grs.On("ThingGetByID", mock.Anything, "1234").Once().Return(i, nil)

	// Make request and validate we get back proper response
	e := httpexpect.New(t, server.URL)
	e.GET("/api/things/1234").Expect().Status(http.StatusOK).JSON().Object().Equal(&i)

	// Check remaining expectations
	grs.AssertExpectations(t)

}

func TestThingDeleteByID(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	grs := new(mocks.GRStore)
	err := Setup(r, grs)
	assert.Nil(t, err)

	// Mock call to item store
	grs.On("ThingDeleteByID", mock.Anything, "1234").Once().Return(nil)

	// Make request and validate we get back proper response
	e := httpexpect.New(t, server.URL)
	e.DELETE("/api/things/1234").Expect().Status(http.StatusNoContent)

	// Check remaining expectations
	grs.AssertExpectations(t)

}
*/
