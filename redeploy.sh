#!/bin/bash

# MulticamObserver - Redeploy Script
# Pulls latest changes from devv and runs a fresh deployment

set -e

echo "══════════════════════════════════════════════"
echo " 🔄 MULTICAM OBSERVER REDEPLOYMENT"
echo "══════════════════════════════════════════════"

# 1. Reset local changes to ensure clean state
echo "🧹 Resetting local changes..."
git reset --hard HEAD

# 2. Pull latest from devv
echo "📥 Pulling latest changes from branch devv..."
git pull origin devv

# 3. Take down current containers
echo "🐳 Stopping current containers..."
docker compose down

# 4. Run the standard deployment
echo "🚀 Running deployment script..."
bash deploy.sh
