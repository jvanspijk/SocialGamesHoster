package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

// The host always listens on every local interface. These implementation
// details were previously exposed as user-configurable settings.
func init() {
	pbmigrations.Register(func(app core.App) error {
		return removeHostNetworkSelectionFields(app)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("host_settings")
		if err != nil {
			return err
		}
		collection.Fields.Add(
			&core.TextField{Name: "bind_address", Max: 45},
			&core.TextField{Name: "preferred_adapter", Max: 200},
		)
		return app.Save(collection)
	}, "1710000014_host_all_local_networks.go")
}

func removeHostNetworkSelectionFields(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("host_settings")
	if err != nil {
		return err
	}
	collection.Fields.RemoveByName("bind_address")
	collection.Fields.RemoveByName("preferred_adapter")
	return app.Save(collection)
}
