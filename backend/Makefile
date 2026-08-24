.PHONY: help setup-owner seed-db migrate test-api clean

help: ## Show this help message
	@echo 'PartFlow Backend - Available Commands:'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ''

setup-owner: ## Create or update owner account
	@echo "🔧 Setting up owner account..."
	@go run scripts/create_owner.go

seed-db: ## Seed database with initial data
	@echo "🌱 Seeding database with initial data..."
	@go run scripts/seed_data.go

migrate: ## Run database migrations
	@echo "🔄 Running database migrations..."
	@psql $(DATABASE_URL) -f migrations/001_fix_users_schema.sql

init: migrate seed-db setup-owner ## Initialize database (migrate + seed + setup owner)
	@echo "✅ Database initialization complete!"

test-api: ## Test API registration
	@echo "🧪 Testing API registration..."
	@curl -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email": "test@partflow.com", "password": "Test123456", "first_name": "Test", "last_name": "User", "phone": "+970599000000"}'

clean: ## Clean temporary files
	@echo "🧹 Cleaning temporary files..."
	@rm -f hash_password.go
	@echo "✅ Clean complete"