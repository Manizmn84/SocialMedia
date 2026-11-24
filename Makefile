include .env

MIGRATIONS_PATH = ./cmd/migrate/migrations

# Support positional arguments as well as NAME=... and STEPS=... .
MIGRATION_COMMAND := $(firstword $(MAKECMDGOALS))
ifneq ($(filter migration-create migrate-create migrate-down,$(MIGRATION_COMMAND)),)
MIGRATION_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
ifneq ($(strip $(MIGRATION_ARGS)),)
$(foreach arg,$(MIGRATION_ARGS),$(eval $(arg):;@:))
endif
endif

.PHONY: migration-create migrate-create migrate-up migrate-down
migration-create migrate-create:
	@test -n "$(or $(NAME),$(MIGRATION_ARGS))" || { echo 'Usage: make migration-create NAME=create_posts'; exit 1; }
	@migrate create -seq -ext sql -dir "$(MIGRATIONS_PATH)" $(or $(NAME),$(MIGRATION_ARGS))

.PHONY: migrate-up
migrate-up:
	@migrate -path="$(MIGRATIONS_PATH)" -database="$(DB_MIGRATOR_ADDR)" up

.PHONY: migrate-down
migrate-down:
	@migrate -path="$(MIGRATIONS_PATH)" -database="$(DB_MIGRATOR_ADDR)" down $(or $(STEPS),$(MIGRATION_ARGS))

