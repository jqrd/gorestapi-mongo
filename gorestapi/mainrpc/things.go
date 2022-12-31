package mainrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jqrd/gorestapi-mongo/gorestapi"
	"github.com/jqrd/gorestapi-mongo/model/db"
	"github.com/jqrd/gorestapi-mongo/model/svc"
	"github.com/jqrd/gorestapi-mongo/pkg/server/render"
	"github.com/jqrd/gorestapi-mongo/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/bson"
)

type ThingsAPI struct {
	s *Server
}

func (s *Server) NewThingsAPI() *ThingsAPI {
	return &ThingsAPI{s}
}

// Create saves a new thing to the database.
// If the thing has new widgets (determined by comparing the name of the
// widgets already in the system), new Widgets are also inserted. Otherwise,
// the IDs of existing widgets are used.
//
// (This API method illustrates why the split between the DB types and SVC types.
// The logic is not complete, as it doesn't take into account differences in Type
// between existing and new widgets)
//
// @ID ThingsAPI.Create
// @Tags things
// @Summary Create a thing
// @Description Create a thing
// @Param thing body svc.Thing true "Thing"
// @Success 200 {object} svc.Thing
// @Failure 400 {object} render.ErrResponse "Invalid Argument"
// @Failure 500 {object} render.ErrResponse "Internal Error"
// @Router /things [post]
func (api *ThingsAPI) Create() http.HandlerFunc {
	action := "ThingsAPI.Create"
	internalError := func(ctx context.Context, w http.ResponseWriter, err error, message string) {
		requestID := middleware.GetReqID(ctx)
		render.ErrInternalWithRequestID(w, requestID, nil)
		api.s.logger.Errorw(fmt.Sprintf("%v error: %v", action, message), "error", err, "request_id", requestID)
	}
	// TODO test
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var thing = new(svc.Thing)
		if err := render.DecodeJSON(r.Body, thing); err != nil {
			render.ErrInvalidRequest(w, err)
			return
		}

		dbThing := &db.Thing{
			Name:        thing.Name,
			Description: thing.Description,
			WidgetIDs:   make([]string, 0, len(thing.Widgets)),
		}

		if len(thing.Widgets) > 0 {
			// TODO cache widgets

			// find existing widgets, insert if not found
			names := make([]string, len(thing.Widgets))
			for i := 0; i < len(thing.Widgets); i++ {
				thing.Widgets[i].EnsureCanonicalName()
				names[i] = thing.Widgets[i].Name
			}
			filter := bson.M{"name": bson.M{"$in": names}}
			existingWidgets, err := api.s.store.Widgets().Find(ctx, filter)
			if err != nil {
				internalError(ctx, w, err, "unable to query widgets")
				return
			}
			existingWidgetsByName := make(map[string]*db.Widget)
			for _, existingWidget := range existingWidgets {
				existingWidgetsByName[existingWidget.Name] = existingWidget
			}
			if len(thing.Widgets) > len(existingWidgets) {
				newWidgets := make([]*db.Widget, 0, len(thing.Widgets))
				for _, widget := range thing.Widgets {
					newWidget := &db.Widget{
						Name: widget.Name,
						Type: widget.Type,
					}
					newWidgets = append(newWidgets, newWidget)
				}
				err := api.s.store.Widgets().InsertMany(ctx, newWidgets)
				if err != nil {
					internalError(ctx, w, err, "unable to insert widgets")
					return
				}
				for _, newWidget := range newWidgets {
					existingWidgetsByName[newWidget.Name] = newWidget
				}
			}
			for i, thingWidget := range thing.Widgets {
				id := existingWidgetsByName[thingWidget.Name].Id
				dbThing.WidgetIDs[i] = id
				thingWidget.WidgetId = id
			}
		}

		err := api.s.store.Things().InsertOne(ctx, dbThing)
		if err != nil {
			if serr, ok := err.(*store.Error); ok {
				render.ErrInvalidRequest(w, serr.ErrorForOp(store.ErrorOpSave))
			} else {
				internalError(ctx, w, err, "unable to insert new thing")
			}
			return
		}

		thing.Id = dbThing.Id

		render.JSON(w, http.StatusOK, thing)
	}
}

// GetById fetches a thing from the database
//
// @ID ThingsAPI.GetByID
// @Tags things
// @Summary Get thing
// @Description Get a thing
// @Param id path string true "ID"
// @Success 200 {object} svc.Thing
// @Failure 400 {object} render.ErrResponse "Invalid Argument"
// @Failure 404 {object} render.ErrResponse "Not Found"
// @Failure 500 {object} render.ErrResponse "Internal Error"
// @Router /things/{id} [get]
func (api *ThingsAPI) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")

		dbThing, err := api.s.store.Things().FindOne(ctx, id)
		if err != nil {
			if err == store.ErrNotFound {
				render.ErrResourceNotFound(w, "thing")
			} else if serr, ok := err.(*store.Error); ok {
				render.ErrInvalidRequest(w, serr.ErrorForOp(store.ErrorOpGet))
			} else {
				requestID := middleware.GetReqID(ctx)
				render.ErrInternalWithRequestID(w, requestID, nil)
				api.s.logger.Errorw("ThingsAPI.GetByID error", "error", err, "request_id", requestID)
			}
			return
		}

		thing, err := toSvcThing(dbThing, api.s.store, ctx)
		if err != nil {
			requestID := middleware.GetReqID(ctx)
			render.ErrInternalWithRequestID(w, requestID, nil)
			api.s.logger.Errorw("ThingsAPI.GetByID error", "error", err, "request_id", requestID)
			return
		}
		render.JSON(w, http.StatusOK, thing)
	}
}

// DeleteByID deleted a thing from the database
//
// @ID ThingsAPI.DeleteByID
// @Tags things
// @Summary Delete thing
// @Description Delete a thing
// @Param id path string true "ID"
// @Success 204 "Success"
// @Failure 400 {object} render.ErrResponse "Invalid Argument"
// @Failure 404 {object} render.ErrResponse "Not Found"
// @Failure 500 {object} render.ErrResponse "Internal Error"
// @Router /things/{id} [delete]
func (api *ThingsAPI) DeleteByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")

		err := api.s.store.Things().DeleteOne(ctx, id)
		if err != nil {
			if err == store.ErrNotFound {
				render.ErrResourceNotFound(w, "thing")
			} else if serr, ok := err.(*store.Error); ok {
				render.ErrInvalidRequest(w, serr.ErrorForOp(store.ErrorOpGet))
			} else {
				requestID := middleware.GetReqID(ctx)
				render.ErrInternalWithRequestID(w, requestID, nil)
				api.s.logger.Errorw("ThingsAPI.GetByID error", "error", err, "request_id", requestID)
			}
			return
		}

		render.NoContent(w)
	}
}

// Find finds things
//
// @ID Find
// @Tags things
// @Summary Find things
// @Description Find things
// @Param q query string false "query"
// @Param offset query int false "offset"
// @Param limit query int false "limit"
// @Success 200 {array} svc.Thing
// @Failure 400 {object} render.ErrResponse "Invalid Argument"
// @Failure 500 {object} render.ErrResponse "Internal Error"
// @Router /things [get]
func (api *ThingsAPI) Find() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		filter := bson.M{}

		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			if err != nil {
				render.ErrInvalidRequest(w, err)
				return
			}
		}

		if query["q"] != nil && len(query["q"]) > 0 && len(query["q"][0]) > 0 {
			unescapedQuery, err := url.QueryUnescape(query["q"][0])
			if err != nil {
				render.ErrInvalidRequest(w, err)
				return
			}
			err = json.Unmarshal([]byte(unescapedQuery), &filter)
			if err != nil {
				render.ErrInvalidRequest(w, err)
				return
			}
		}

		// TODO offset and limit
		dbThings, err := api.s.store.Things().Find(ctx, filter)
		if err != nil {
			if err == store.ErrNotFound {
				render.ErrResourceNotFound(w, "thing")
			} else if serr, ok := err.(*store.Error); ok {
				render.ErrInvalidRequest(w, serr.ErrorForOp(store.ErrorOpGet))
			} else {
				requestID := middleware.GetReqID(ctx)
				render.ErrInternalWithRequestID(w, requestID, nil)
				api.s.logger.Errorw("ThingsAPI.GetByID error", "error", err, "request_id", requestID)
			}
			return
		}

		svcThings := make([]*svc.Thing, len(dbThings))
		for i := 0; i < len(dbThings); i++ {
			svcThings[i], err = toSvcThing(dbThings[i], api.s.store, ctx)
			if err != nil {
				requestID := middleware.GetReqID(ctx)
				render.ErrInternalWithRequestID(w, requestID, nil)
				api.s.logger.Errorw("ThingsAPI.GetByID error", "error", err, "request_id", requestID)
				return
			}
		}
		render.JSON(w, http.StatusOK, svcThings)
	}
}

func toSvcThing(dbThing *db.Thing, store gorestapi.DataStore, ctx context.Context) (*svc.Thing, error) {
	thing := &svc.Thing{
		Id:          dbThing.Id,
		Name:        dbThing.Name,
		Description: dbThing.Description,
		Widgets:     make([]*svc.ThingWidget, 0, len(dbThing.WidgetIDs)),
	}

	filter := bson.M{"_id": bson.M{"$in": dbThing.WidgetIDs}}
	widgets, err := store.Widgets().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	widgetsById := make(map[string]*db.Widget)
	for _, widget := range widgets {
		widgetsById[widget.Id] = widget
	}
	for i, id := range dbThing.WidgetIDs {
		dbWidget := widgetsById[id]
		thing.Widgets[i] = &svc.ThingWidget{
			WidgetId: id,
			Name:     dbWidget.Name,
			Type:     dbWidget.Type,
		}
	}

	return thing, nil
}
