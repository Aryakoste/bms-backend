# Building Management System - Multi-Society Backend

## 🏢 SOCIETY-BASED ACCESS CONTROL SYSTEM

This backend now supports **multiple societies** with **complete data segregation** using society access codes!

## 🌟 New Society Features

### 🔐 Society Access Flow
1. **Society Code Validation** - Users enter society code first
2. **Role Selection** - Choose Secretary/Resident/Security
3. **Society-Scoped Login** - Authentication within specific society
4. **Data Segregation** - All data operations are society-aware

### 🏢 Multi-Society Support
- **Multiple societies** can use the same backend
- **Complete data isolation** between societies
- **Society-specific** users, visitors, amenities, etc.
- **Society access codes** for security

## 🚀 Quick Start

### 1. Start MongoDB
```bash
docker run -d -p 27017:27017 --name mongodb mongo:7.0
```

### 2. Setup Backend
```bash
cd bms-backend-society
go mod tidy
go run scripts/seed.go      # Creates 3 demo societies
go run cmd/server/main.go   # Starts multi-society API
```

## 🏢 Demo Societies

### Society 1: GREEN001 (Green Valley Apartments)
- **Secretary**: rajesh@demo.com / demo123
- **Resident**: priya@demo.com / demo123
- **Security**: security@demo.com / demo123

### Society 2: BLUE002 (Blue Hills Society)
- **Secretary**: admin@bluehills.com / demo123
- **Resident**: resident@bluehills.com / demo123
- **Security**: guard@bluehills.com / demo123

### Society 3: SUN003 (Sunrise Residency)
- **Secretary**: admin@SUN003.com / demo123

## 🔄 New Authentication Flow

### 1. Validate Society Code
```bash
curl -X POST http://localhost:8080/api/v1/society/validate   -H "Content-Type: application/json"   -d '{"code":"GREEN001"}'
```

**Response:**
```json
{
  "valid": true,
  "society": {
    "id": "...",
    "name": "Green Valley Apartments",
    "code": "GREEN001",
    "address": "123 Green Valley Road",
    "city": "Mumbai",
    "state": "Maharashtra"
  }
}
```

### 2. Login with Society Code
```bash
curl -X POST http://localhost:8080/api/v1/auth/login   -H "Content-Type: application/json"   -d '{
    "email": "rajesh@demo.com",
    "password": "demo123",
    "society_code": "GREEN001"
  }'
```

**Response:**
```json
{
  "token": "jwt_token_with_society_context",
  "user": { 
    "id": "...",
    "name": "Rajesh Kumar",
    "email": "rajesh@demo.com",
    "role": "secretary",
    "society_code": "GREEN001"
  },
  "society": {
    "id": "...",
    "name": "Green Valley Apartments",
    "code": "GREEN001",
    "address": "123 Green Valley Road"
  }
}
```

### 3. Use JWT Token (Now Includes Society Context)
```bash
curl -X GET http://localhost:8080/api/v1/visitors   -H "Authorization: Bearer JWT_TOKEN"
```

**Only returns visitors for the user's society!**

## 📋 Enhanced APIs (All Society-Aware)

### 🏢 Society Management
- `POST /api/v1/society/validate` - Validate society access code

### 🔐 Authentication (Society-Enhanced)
- `POST /api/v1/auth/login` - Login with society code
- `POST /api/v1/auth/register` - Register with society code
- `GET /api/v1/users/profile` - Get profile (society-scoped)

### 👥 Users (Society-Scoped)
- `GET /api/v1/users/residents` - List residents in same society
- `GET /api/v1/users/stats` - Dashboard stats for society
- `GET /api/v1/users/:id` - Get user in same society

### 👤 Visitors (Society-Scoped)
- All visitor endpoints now filter by society
- QR codes include society code
- Only society members can approve visitors

### 💰 Maintenance (Society-Scoped)
- Maintenance records isolated by society
- Payments processed within society context

### 🏊 Amenities (Society-Scoped)
- Each society has its own amenities
- Booking conflicts checked within society
- No cross-society amenity access

### 📢 Notices (Society-Scoped)
- Notices isolated by society
- Only society secretaries can manage notices

## 🛡️ Data Security Features

### 🔒 Complete Data Isolation
- **Society-based filtering** on all database queries
- **JWT tokens** include society context
- **API endpoints** automatically filter by society
- **No cross-society data access**

### 🏢 Society Access Control
- **Unique society codes** (GREEN001, BLUE002, etc.)
- **Society validation** before login
- **Society-scoped authentication**
- **Society-aware data operations**

### 🔐 Enhanced Security
- **Compound unique indexes** (email + society_code)
- **Society context** in all protected routes
- **Middleware filtering** by society
- **QR codes** include society identification

## 🧪 Testing Multi-Society System

### Test Society Isolation
```bash
# Login to Society 1
curl -X POST http://localhost:8080/api/v1/auth/login   -H "Content-Type: application/json"   -d '{"email":"rajesh@demo.com","password":"demo123","society_code":"GREEN001"}'

# Login to Society 2  
curl -X POST http://localhost:8080/api/v1/auth/login   -H "Content-Type: application/json"   -d '{"email":"admin@bluehills.com","password":"demo123","society_code":"BLUE002"}'

# Each will get different data sets!
```

### Test Cross-Society Protection
```bash
# Try to access Society 1 data with Society 2 token
curl -X GET http://localhost:8080/api/v1/visitors   -H "Authorization: Bearer SOCIETY2_TOKEN"

# Will only return Society 2 visitors - complete isolation!
```

## 🎯 Key Enhancements

- ✅ **Multi-Society Support** - Multiple societies per backend
- ✅ **Society Access Codes** - Secure society identification  
- ✅ **Complete Data Isolation** - No cross-society data access
- ✅ **Society-Aware Authentication** - JWT includes society context
- ✅ **Enhanced Security** - Society-based data filtering
- ✅ **Flexible Architecture** - Easy to add more societies

## 📁 Enhanced File Structure

```
bms-backend-society/
├── cmd/server/main.go           # Multi-society server
├── api/routes/routes.go         # Society-aware routes
├── internal/
│   ├── handlers/               # All society-aware handlers
│   │   ├── auth_handler.go     # Society validation + auth
│   │   ├── user_handler.go     # Society-scoped users
│   │   ├── visitor_handler.go  # Society-scoped visitors
│   │   ├── maintenance_handler.go # Society-scoped maintenance
│   │   ├── amenity_handler.go  # Society-scoped amenities
│   │   └── notice_handler.go   # Society-scoped notices
│   ├── models/models.go        # Enhanced with Society model
│   ├── middleware/auth.go      # Society context middleware
│   └── utils/utils.go          # Society-aware QR generation
├── scripts/seed.go             # Multi-society sample data
└── README.md                   # This comprehensive guide
```

**This is now a complete MULTI-SOCIETY building management system!** 🏢✨

Each society operates independently with complete data isolation while sharing the same robust backend infrastructure.
