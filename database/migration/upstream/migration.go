package version

import (
	v00 "opentracing-playground/database/migration/upstream/v00"

	"github.com/go-gormigrate/gormigrate/v2"
)

// ModelSchemaList Model Structs
var ModelSchemaList = []*gormigrate.Migration{
	&v00.Init,
}
