package db

import model "github.com/jqrd/gorestapi-mongo/model"

func (w *Widget) EnsureCanonicalName() string {
	w.Name = model.EnsureCanonicalName(w.Name)
	return w.Name
}
