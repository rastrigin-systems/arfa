# Ubik Enterprise - HTML Prototype

Fast HTML+Tailwind+htmx prototype for rapid development and testing.

## 🚀 Quick Start

### 1. Start the database and server

```bash
# From pivot directory
make db-up
make db-seed  # Load test data

# Build and run the server
go run cmd/server/main.go
```

### 2. Open in browser

```
http://localhost:3001/
```

This will redirect to the login page.

### 3. Login with test credentials

**Test Accounts (all passwords: `password123`):**
- `alice@acme.com` - Super Admin at Acme Corp
- `bob@acme.com` - Admin at Acme Corp
- `charlie@acme.com` - Developer at Acme Corp

## 📄 Available Pages

- **`/login.html`** - Login page (default)
- **`/dashboard.html`** - Main dashboard with stats
- **`/employees.html`** - Employee list with filters
- **`/teams.html`** - Teams management (TODO)
- **`/agents.html`** - Agent catalog (TODO)
- **`/settings.html`** - Organization settings (TODO)

## 🛠 Technology Stack

- **HTML5** - Simple, fast, no build step
- **Tailwind CSS** - Via CDN for rapid styling
- **htmx** - AJAX without JavaScript
- **Alpine.js** - Minimal JavaScript for interactivity
- **Go Chi Router** - Serves static files + API

## ✨ Features

### ✅ Implemented
- Login with JWT authentication
- Dashboard with stats (employee count, teams, mock usage)
- Employee list with pagination
- Real API integration (GET /employees, GET /teams, GET /agents)
- Responsive design
- Token-based auth with localStorage
- Auto-redirect to login if not authenticated

### 🚧 Partial
- Dashboard analytics (some data mocked)
- Filters on employee list (status filter works)

### 📋 TODO
- Employee detail page
- Create/Edit employee forms
- Teams list and detail pages
- Agent catalog page
- Organization settings page
- More wireframe conversions

## 🔌 API Integration

All pages connect to the real backend API at `/api/v1`:

```javascript
// Auth stored in localStorage
const token = localStorage.getItem('auth_token');

// API calls with auth
fetch('/api/v1/employees', {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});
```

### Endpoints Used

| Page | Endpoints |
|------|-----------|
| Login | `POST /api/v1/auth/login` |
| Dashboard | `GET /auth/me`, `/organizations/current`, `/employees`, `/teams`, `/agents` |
| Employees | `GET /employees?page=1&per_page=20&status=active` |

## 🎨 Adding New Pages

### Quick Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Page Title - Ubik Enterprise</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.13.0/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-50" x-data="yourPageData()">
    <!-- Copy header/nav from another page -->

    <!-- Your content here -->

    <script>
        // Copy auth helpers from another page

        function yourPageData() {
            return {
                async init() {
                    await loadCurrentUser();
                    // Load your data
                }
            }
        }
    </script>
</body>
</html>
```

### Register Route in Go

Edit `cmd/server/main.go`:

```go
router.Get("/yourpage.html", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./static/yourpage.html")
})
```

Restart server: `go run cmd/server/main.go`

## 🧪 Testing

### 1. Manual Testing

1. Start server: `go run cmd/server/main.go`
2. Open browser: `http://localhost:3001/`
3. Login with test account
4. Navigate through pages
5. Check browser console for errors

### 2. API Testing

```bash
# Health check
curl http://localhost:3001/api/v1/health

# Login
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@acme.com","password":"password123"}'

# Get employees (with token)
curl http://localhost:3001/api/v1/employees \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 📐 Design System

### Colors
- **Primary**: Blue (#3b82f6)
- **Success**: Green (#10b981)
- **Warning**: Yellow (#f59e0b)
- **Danger**: Red (#ef4444)
- **Gray scale**: Tailwind defaults

### Components
- **Buttons**: `bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md`
- **Cards**: `bg-white rounded-lg shadow border border-gray-200 p-6`
- **Tables**: Striped rows with hover effects
- **Forms**: Standard Tailwind form classes

## 🚀 Performance

- **No build step** - Edit and refresh
- **CDN resources** - Cached by browser
- **Minimal JavaScript** - Only Alpine.js for interactivity
- **Direct API calls** - No extra abstraction layers

## 🔄 Migration Path

When ready to move to production:

1. **Option A: Keep HTML+htmx** (if it works well)
   - Move to proper asset pipeline
   - Add template rendering (Go templates)
   - Bundle CSS/JS

2. **Option B: Migrate to Next.js**
   - Convert pages to React components
   - Keep same API calls
   - Use shadcn/ui for components
   - Reuse same design system

## 📝 Notes

- **Tokens stored in localStorage** - Not httpOnly cookies (prototype only)
- **No form validation** - Basic browser validation only
- **Limited error handling** - Shows alerts for now
- **Mock data** - Some dashboard stats are hardcoded

## 🐛 Known Issues

1. **Activity logs** - Not exposed via API yet (dashboard shows mock data)
2. **Team filter** - Employee list team filter not implemented in API
3. **Usage stats** - No aggregation endpoint yet
4. **No routing** - Manual URL changes, no SPA routing

## 🎯 Next Steps

1. ✅ Login page
2. ✅ Dashboard page
3. ✅ Employee list page
4. ⏳ Employee detail page
5. ⏳ Teams list and detail
6. ⏳ Agent catalog and configs
7. ⏳ Organization settings

## 📚 Resources

- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [htmx Docs](https://htmx.org/docs/)
- [Alpine.js Docs](https://alpinejs.dev/start-here)
- [Wireframes](../wireframes/) - Original ASCII wireframes

---

**Last Updated:** 2025-10-29
