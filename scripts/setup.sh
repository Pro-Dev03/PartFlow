#!/bin/bash

# PartFlow Setup Script
# This script helps set up the development environment

set -e

echo "🚀 Setting up PartFlow Development Environment..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

echo "✅ All prerequisites are installed."

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
    echo "⚠️  Please edit .env file with your configuration before proceeding."
else
    echo "✅ .env file already exists."
fi

# Install backend dependencies
echo "📦 Installing backend dependencies..."
cd backend
go mod download
cd ..

# Install frontend dependencies
echo "📦 Installing frontend dependencies..."
cd frontend
npm install
cd ..

# Start Docker services
echo "🐳 Starting Docker services..."
docker-compose up -d postgres redis

echo "✅ Setup completed successfully!"
echo ""
echo "📝 Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run backend: cd backend && go run cmd/api/main.go"
echo "3. Run frontend: cd frontend && npm run dev"
echo "4. Or use docker-compose: docker-compose up"
