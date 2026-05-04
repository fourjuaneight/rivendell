# Migrations

PocketBase tracks schema changes via Go migration files in `migrations/`. Each migration runs exactly once per environment, recorded in a `_migrations` SQLite table.

## How it works

Migrations are registered via `m.Register(up, down)` inside `init()` functions. Because `main.go` blank-imports the `migrations` package, all `init()` functions run at startup — registering every migration before the app boots. PocketBase then applies any that haven't been recorded in `_migrations` yet.

With `Automigrate: true` in `main.go`, this happens automatically on every `serve` startup. No manual `migrate up` command needed.

## File naming

```
migrations/{unix_timestamp}_{description}.go
```

Timestamp determines execution order. Use the current Unix time when creating a new file.

```sh
date +%s  # get current Unix timestamp
```

## Writing a migration

```go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        // up: apply the change
        collection, err := app.FindCollectionByNameOrId("my_collection")
        if err != nil {
            return err
        }

        collection.Fields.Add(&core.TextField{Name: "new_field"})
        return app.Save(collection)
    }, func(app core.App) error {
        // down: revert the change (pass nil if rollback isn't possible)
        collection, err := app.FindCollectionByNameOrId("my_collection")
        if err != nil {
            return err
        }

        collection.Fields.RemoveByName("new_field")
        return app.Save(collection)
    })
}
```

Use `app.RunInTransaction()` when saving multiple collections so a partial failure rolls back cleanly:

```go
return app.RunInTransaction(func(txApp core.App) error {
    for _, collection := range collections {
        if err := txApp.Save(collection); err != nil {
            return err
        }
    }
    return nil
})
```

## Schema package

Collection definitions live in `schema/collections.go`, not inline in migration files. This keeps the schema readable and lets multiple migrations reference the same builders.

When modifying an existing collection, write a new migration that fetches the collection and applies only the delta — don't edit the original migration file, since it has already run on existing environments.

## Common operations

**Add a field:**
```go
collection, _ := app.FindCollectionByNameOrId("bookmarks")
collection.Fields.Add(&core.TextField{Name: "source"})
app.Save(collection)
```

**Remove a field:**
```go
collection, _ := app.FindCollectionByNameOrId("bookmarks")
collection.Fields.RemoveByName("source")
app.Save(collection)
```

**Drop a collection (idempotent):**
```go
collection, err := app.FindCollectionByNameOrId("old_table")
if errors.Is(err, sql.ErrNoRows) {
    return nil // already gone
}
app.Delete(collection)
```

## Running manually

Though `Automigrate` handles this on startup, you can also trigger migrations explicitly:

```sh
./rivendell migrate up       # apply all pending
./rivendell migrate down     # revert last applied
./rivendell migrate history  # list applied migrations
```

Or with `go run`:
```sh
go run . migrate up
```
