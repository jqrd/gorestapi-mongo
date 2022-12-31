package store

import (
	mongodriver "github.com/jqrd/gorestapi-mongo/store/driver/mongo"
)

func TryTranslateMongoError(err error) error {
	if err == nil {
		return nil
	}

	switch mongodriver.GetKnownError(err) {
	case mongodriver.Error_NoDocumentsFound:
		return ErrNotFound
	default:
		return err
	}
}
