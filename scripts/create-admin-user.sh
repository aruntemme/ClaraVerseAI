#!/bin/bash

# ============================================
# ClaraVerse - Create Default Admin User
# ============================================
# Creates a default admin user if one doesn't exist
#
# Default credentials:
#   Email: admin@localhost
#   Password: admin
#
# ⚠️  Change password on first login!
# ============================================

set -e

echo "🔐 ClaraVerse - Creating Default Admin User"
echo "==========================================="
echo ""

# Check if MongoDB is accessible
MONGODB_URI="${MONGODB_URI:-mongodb://localhost:27017/claraverse}"

echo "📡 Checking MongoDB connection..."
if ! mongosh "$MONGODB_URI" --eval "db.version()" --quiet > /dev/null 2>&1; then
    echo "❌ Cannot connect to MongoDB at: $MONGODB_URI"
    echo "   Make sure MongoDB is running:"
    echo "   docker-compose up -d mongodb"
    exit 1
fi

echo "✅ MongoDB connected"
echo ""

# Check if admin user already exists
echo "🔍 Checking for existing admin user..."
ADMIN_EXISTS=$(mongosh "$MONGODB_URI" --quiet --eval "db.users.countDocuments({ email: 'admin@localhost' })")

if [ "$ADMIN_EXISTS" != "0" ]; then
    echo "✅ Admin user already exists"
    echo ""
    echo "To reset admin password, delete the user first:"
    echo "  mongosh '$MONGODB_URI' --eval \"db.users.deleteOne({ email: 'admin@localhost' })\""
    exit 0
fi

echo "📝 Creating default admin user..."
echo ""

# Generate password hash (you'll need to do this with your JWT auth in the backend)
# For now, we'll create the user and set a flag to require password change
mongosh "$MONGODB_URI" --quiet --eval "
db.users.insertOne({
    email: 'admin@localhost',
    emailVerified: true,
    role: 'admin',
    subscriptionTier: 'pro',
    subscriptionStatus: 'active',
    refreshTokenVersion: 0,
    requirePasswordChange: true,
    createdAt: new Date(),
    updatedAt: new Date(),
    _note: 'Default admin user - password must be set via registration endpoint'
})
"

echo "✅ Created default admin user: admin@localhost"
echo ""
echo "==========================================="
echo "⚠️  IMPORTANT: Complete Setup Required"
echo "==========================================="
echo ""
echo "The admin user has been created but needs a password."
echo "Please complete setup by:"
echo ""
echo "1. Start the backend server:"
echo "   docker-compose up"
echo ""
echo "2. Register the admin user via the API:"
echo "   curl -X POST http://localhost:3001/api/auth/register \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"email\":\"admin@localhost\",\"password\":\"your-secure-password\"}'"
echo ""
echo "OR use the frontend to register with email: admin@localhost"
echo ""
echo "🔒 Make sure to use a strong password!"
echo ""
