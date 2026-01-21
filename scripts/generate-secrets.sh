#!/bin/bash

# ============================================
# ClaraVerse Secret Generation Script
# ============================================
# Generates cryptographically secure secrets for .env file
#
# Usage:
#   ./scripts/generate-secrets.sh
#
# This script will:
#   1. Generate ENCRYPTION_MASTER_KEY (32 bytes)
#   2. Generate JWT_SECRET (64 bytes)
#   3. Update .env file with generated secrets
# ============================================

set -e  # Exit on error

echo "🔐 ClaraVerse Secret Generation Script"
echo "======================================"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.minimal..."
    cp .env.minimal .env
    echo "✅ Created .env file"
else
    echo "✅ Found existing .env file"
fi

echo ""
echo "🎲 Generating cryptographically secure secrets..."
echo ""

# Generate ENCRYPTION_MASTER_KEY (32 bytes = 64 hex characters)
ENCRYPTION_KEY=$(openssl rand -hex 32)
echo "✅ Generated ENCRYPTION_MASTER_KEY (32 bytes)"

# Generate JWT_SECRET (64 bytes = 128 hex characters)
JWT_SECRET=$(openssl rand -hex 64)
echo "✅ Generated JWT_SECRET (64 bytes)"

echo ""
echo "📝 Updating .env file with generated secrets..."

# Update .env file (works on both macOS and Linux)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/ENCRYPTION_MASTER_KEY=.*/ENCRYPTION_MASTER_KEY=$ENCRYPTION_KEY/" .env
    sed -i '' "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
else
    # Linux
    sed -i "s/ENCRYPTION_MASTER_KEY=.*/ENCRYPTION_MASTER_KEY=$ENCRYPTION_KEY/" .env
    sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
fi

echo "✅ Updated .env file"
echo ""
echo "=========================================="
echo "✅ Secret generation complete!"
echo "=========================================="
echo ""
echo "Your .env file now contains:"
echo "  • ENCRYPTION_MASTER_KEY: $ENCRYPTION_KEY"
echo "  • JWT_SECRET: ${JWT_SECRET:0:32}..."
echo ""
echo "⚠️  IMPORTANT: Keep these secrets secure!"
echo "   • Never commit .env to version control"
echo "   • Losing ENCRYPTION_MASTER_KEY = losing encrypted data"
echo ""
echo "Next steps:"
echo "  1. Review your .env file: cat .env"
echo "  2. Start services: docker-compose up"
echo "  3. Access at: http://localhost:5173"
echo ""
echo "Default admin account:"
echo "  Email: admin@localhost"
echo "  Password: admin"
echo "  (Change on first login!)"
echo ""
