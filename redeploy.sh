#!/bin/bash

# MulticamObserver - Redeploy Script
# Pulls latest changes from main and runs a fresh deployment

set -e

echo "══════════════════════════════════════════════"
echo " 🔄 MULTICAM OBSERVER REDEPLOYMENT"
echo "══════════════════════════════════════════════"

# 1. Reset local changes to ensure clean state
echo "🧹 Resetting local changes..."
git reset --hard HEAD

# 2. Pull latest from main
echo "📥 Pulling latest changes from branch main..."
git pull origin main

# 3. Take down current containers and clear volumes
echo "🐳 Stopping current containers and clearing old volumes..."
docker compose down -v

# 4. Run the standard deployment
echo "🚀 Running deployment script..."
bash deploy.sh
