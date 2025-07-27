ifneq (,$(wildcard .env))
    include .env
    export
endif

MIGRATION_DIR?=./migrations

MIGRATE=migrate -source file://$(MIGRATION_DIR) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

migration_create:
	migrate create -seq -dir $(MIGRATION_DIR) -ext sql $(name)

migration_up:
	$(MIGRATE) up

migration_down:
	$(MIGRATE) down 1

migration_drop:
	$(MIGRATE) drop -f