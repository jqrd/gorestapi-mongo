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
	"github.com/jqrd/gorestapi-mongo/model"
	"github.com/jqrd/gorestapi-mongo/model/db"
	"github.com/jqrd/gorestapi-mongo/model/svc"
	"github.com/jqrd/gorestapi-mongo/pkg/test"
	"github.com/jqrd/gorestapi-mongo/pkg/test/matchbyhelper"
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
	input := &svc.Thing{
		Name:        "name",
		Description: "descr",
		Widgets:     []*svc.ThingWidget{},
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
	e.POST("/api/things").WithJSON(input).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Equal(expected)

	// Check remaining expectations
	store.AssertExpectations(t)
	thingsCol.AssertExpectations(t)
}
func TestThingPostWithWidgetCreation(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	store := new(mocks.DataStore)
	err := Setup(r, store)
	assert.Nil(t, err)

	// Create Item
	input := &svc.Thing{
		Name:        "name",
		Description: "descr",
		Widgets: []*svc.ThingWidget{
			{Name: "widget1", Type: model.WidgetType_Type1},
			{Name: "widget2", Type: model.WidgetType_Special},
		},
	}

	expected := test.CloneSvcThing(input)
	expected.Id = "a42"
	expected.Widgets[0].WidgetId = "b456"
	expected.Widgets[1].WidgetId = "c123"

	// Mock calls to data store
	m := matchbyhelper.New()
	thingsCol := new(mocks.MongoCollection[*db.Thing])
	store.On("Things").Return(thingsCol)
	widgetsCol := new(mocks.MongoCollection[*db.Widget])
	store.On("Widgets").Return(widgetsCol)

	thingsCol.On("InsertOne", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			args.Get(1).(*db.Thing).Id = expected.Id
		}).
		Return(nil)

	dbWidgets := []*db.Widget{
		{Id: expected.Widgets[1].WidgetId, Name: input.Widgets[1].Name, Type: input.Widgets[1].Type},
	}
	expectedFilter := bson.M{"name": bson.M{"$in": []string{input.Widgets[0].Name, input.Widgets[1].Name}}}
	widgetsCol.On("Find", mock.Anything, matchbyhelper.MockMatchedBy(t, m, test.MatchJson(t, expectedFilter))).
		Return(dbWidgets, nil)

	expectedInsertList := []*db.Widget{
		{Name: input.Widgets[0].Name, Type: input.Widgets[0].Type},
	}
	widgetsCol.On("InsertMany", mock.Anything, matchbyhelper.MockMatchedBy(t, m, test.MatchJson(t, expectedInsertList))).
		Run(func(args mock.Arguments) {
			args.Get(1).([]*db.Widget)[0].Id = expected.Widgets[0].WidgetId
		}).
		Return(nil)

	// Make request and validate we get back proper response
	e := httpexpect.Default(t, server.URL)
	res := e.POST("/api/things").WithJSON(input).
		Expect()
	res.Status(http.StatusOK).
		JSON().Object().Equal(expected)

	// Check remaining expectations
	m.BeginAssert()
	store.AssertExpectations(t)
	thingsCol.AssertExpectations(t)
	widgetsCol.AssertExpectations(t)
}

func TestThingGetByID(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	store := new(mocks.DataStore)
	err := Setup(r, store)
	assert.Nil(t, err)

	// Create Item
	dbThing := &db.Thing{
		Id:        "1234",
		Name:      "name",
		WidgetIDs: []string{"1", "2"},
	}

	dbWidgets := []*db.Widget{
		{Id: "1", Name: "name 1", Type: model.WidgetType_Type1, Description: "111"},
		{Id: "2", Name: "name 2", Type: model.WidgetType_Type2, Description: "222"},
	}

	// Mock call to item store
	thingsCol := new(mocks.MongoCollection[*db.Thing])
	store.On("Things").Return(thingsCol)
	widgetsCol := new(mocks.MongoCollection[*db.Widget])
	store.On("Widgets").Return(widgetsCol)

	thingsCol.On("FindOne", mock.Anything, dbThing.Id).
		Once().Return(dbThing, nil)

	widgetsCol.On("Find", mock.Anything, bson.M{"_id": bson.M{"$in": dbThing.WidgetIDs}}).
		Once().Return(dbWidgets, nil)

	// Make request and validate we get back proper response
	expected := &svc.Thing{
		Id:          dbThing.Id,
		Name:        dbThing.Name,
		Description: dbThing.Description,
		Widgets: []*svc.ThingWidget{
			{WidgetId: "1", Name: "name 1", Type: model.WidgetType_Type1},
			{WidgetId: "2", Name: "name 2", Type: model.WidgetType_Type2},
		},
	}
	e := httpexpect.Default(t, server.URL)
	e.GET("/api/things/1234").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Equal(expected)

	// Check remaining expectations
	store.AssertExpectations(t)
	thingsCol.AssertExpectations(t)
	widgetsCol.AssertExpectations(t)
}

func TestThingsFind(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	store := new(mocks.DataStore)
	err := Setup(r, store)
	assert.Nil(t, err)

	// Return Item
	dbThings := []*db.Thing{
		{Id: "id1", Name: "thing 1"},
		{Id: "id2", Name: "something"},
	}
	expected := []*svc.Thing{
		{Id: "id1", Name: "thing 1", Widgets: []*svc.ThingWidget{}},
		{Id: "id2", Name: "something", Widgets: []*svc.ThingWidget{}},
	}

	// Mock call to item store
	thingsCol := new(mocks.MongoCollection[*db.Thing])
	store.On("Things").Return(thingsCol)

	filter := map[string]interface{}{"name": map[string]interface{}{"$regex": "thing"}}
	thingsCol.On("Find", mock.Anything, mock.MatchedBy(test.MatchJson(t, filter))).Once().Return(dbThings, nil)

	// Make request and validate we get back proper response
	e := httpexpect.Default(t, server.URL)
	filterString, _ := test.JSON(filter)
	e.GET("/api/things").WithQuery("q", filterString).
		Expect().
		Status(http.StatusOK).
		JSON().Array().Equal(expected)

	// Check remaining expectations
	store.AssertExpectations(t)
	thingsCol.AssertExpectations(t)
}

func TestThingDeleteByID(t *testing.T) {

	// Create test server
	r := chi.NewRouter()
	server := httptest.NewServer(r)
	defer server.Close()

	// Mock Store and server
	store := new(mocks.DataStore)
	err := Setup(r, store)
	assert.Nil(t, err)

	// Mock call to item store
	thingsCol := new(mocks.MongoCollection[*db.Thing])
	store.On("Things").Return(thingsCol)
	thingsCol.On("DeleteOne", mock.Anything, "1234").Once().Return(nil)

	// Make request and validate we get back proper response
	e := httpexpect.Default(t, server.URL)
	e.DELETE("/api/things/1234").
		Expect().
		Status(http.StatusNoContent)

	// Check remaining expectations
	store.AssertExpectations(t)

}
