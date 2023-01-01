package test

import (
	"encoding/json"

	"github.com/jqrd/gorestapi-mongo/model/svc"
)

func CloneViaJson[T interface{}](val T) T {

	bytes, _ := json.Marshal(val)
	copy := new(T)
	json.Unmarshal(bytes, copy)
	return *copy
}

func CloneSvcWidget(val *svc.Widget) *svc.Widget {
	clone := CloneViaJson(*val)
	return &clone
}

func CloneSvcThing(val *svc.Thing) *svc.Thing {
	clone := CloneViaJson(*val)
	return &clone
}
