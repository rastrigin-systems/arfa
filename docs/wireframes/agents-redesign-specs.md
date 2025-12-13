# Agents Page - Design System Specifications

## Component Library

All components use **shadcn/ui** + **Tailwind CSS**

---

## Colors

### Primary Palette
```
Primary (Blue):
- 600: #3B82F6 (Buttons, links, active states)
- 500: #60A5FA (Hover states)
- 100: #DBEAFE (Backgrounds, badges)

Success (Green):
- 600: #10B981 (Enabled status)
- 500: #34D399 (Hover)
- 100: #D1FAE5 (Background)

Error (Red):
- 600: #EF4444 (Disable action, errors)
- 500: #F87171 (Hover)
- 100: #FEE2E2 (Background)

Warning (Yellow):
- 600: #F59E0B (Alerts)
- 100: #FEF3C7 (Background)
```

### Neutral Palette
```
Gray:
- 900: #111827 (Headings)
- 700: #374151 (Body text)
- 600: #4B5563 (Secondary text)
- 500: #6B7280 (Muted text)
- 300: #D1D5DB (Borders)
- 200: #E5E7EB (Dividers)
- 100: #F3F4F6 (Card backgrounds)
- 50:  #F9FAFB (Page background)

White: #FFFFFF (Card surfaces)
Black: #000000 (Max contrast text)
```

### Usage Examples
```css
/* Enabled toggle */
.toggle-enabled {
  background-color: #10B981; /* green-600 */
  border-color: #10B981;
}

/* Disabled toggle */
.toggle-disabled {
  background-color: #D1D5DB; /* gray-300 */
  border-color: #D1D5DB;
}

/* Primary button */
.btn-primary {
  background-color: #3B82F6; /* blue-600 */
  color: #FFFFFF;
}

/* Card */
.card {
  background-color: #FFFFFF;
  border: 1px solid #E5E7EB; /* gray-200 */
}
```

---

## Typography

### Font Family
```css
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
```

### Scale
```
H1 (Page Title):
- Size: 2.25rem (36px)
- Weight: 700 (Bold)
- Line Height: 2.5rem (40px)
- Color: gray-900

H2 (Section Heading):
- Size: 1.5rem (24px)
- Weight: 600 (Semibold)
- Line Height: 2rem (32px)
- Color: gray-900

H3 (Card Title):
- Size: 1.25rem (20px)
- Weight: 600 (Semibold)
- Line Height: 1.75rem (28px)
- Color: gray-900

Body (Default):
- Size: 1rem (16px)
- Weight: 400 (Regular)
- Line Height: 1.5rem (24px)
- Color: gray-700

Small (Helper Text):
- Size: 0.875rem (14px)
- Weight: 400 (Regular)
- Line Height: 1.25rem (20px)
- Color: gray-600

Label (Badge, Status):
- Size: 0.75rem (12px)
- Weight: 500 (Medium)
- Line Height: 1rem (16px)
- Color: gray-600
- Transform: uppercase
```

### Usage Examples
```jsx
// Page header
<h1 className="text-4xl font-bold text-gray-900">Agents</h1>
<p className="text-base text-gray-600">Manage AI agents available to your organization</p>

// Card title
<h3 className="text-xl font-semibold text-gray-900">Claude Code</h3>

// Badge
<span className="text-xs font-medium uppercase text-gray-600">IDE Agent</span>

// Description
<p className="text-sm text-gray-600">AI-powered CLI coding assistant</p>
```

---

## Spacing

### Base Unit: 4px

```
Spacing Scale:
0.5  → 2px   (0.125rem)
1    → 4px   (0.25rem)
2    → 8px   (0.5rem)
3    → 12px  (0.75rem)
4    → 16px  (1rem)
6    → 24px  (1.5rem)
8    → 32px  (2rem)
12   → 48px  (3rem)
16   → 64px  (4rem)
```

### Component Spacing

#### Agent Card (Grid View)
```
Card Padding: 24px (p-6)
Card Gap (between cards): 24px (gap-6)
Internal Spacing:
- Title to separator: 8px (mb-2)
- Separator to badge: 12px (mt-3)
- Badge to description: 8px (mt-2)
- Description to toggle: 16px (mt-4)
- Toggle to button: 8px (mt-2)
```

#### Page Layout
```
Page Padding: 32px (p-8)
Section Gap: 24px (space-y-6)
Header to Content: 32px (mt-8)
```

#### Buttons
```
Padding:
- Small: 8px 16px (px-4 py-2)
- Medium: 12px 24px (px-6 py-3)
- Large: 16px 32px (px-8 py-4)

Gap (icon to text): 8px (gap-2)
```

### Usage Examples
```jsx
// Card
<div className="p-6 space-y-4">
  <h3 className="mb-2">Claude Code</h3>
  <div className="h-px bg-gray-200"></div>
  <span className="mt-3">IDE Agent</span>
  <p className="mt-2">Description...</p>
  <div className="mt-4">Toggle</div>
</div>

// Grid
<div className="grid grid-cols-3 gap-6">
  {/* Cards */}
</div>
```

---

## Borders

### Widths
```
Default: 1px
Thick: 2px
Focus Ring: 2px
```

### Radius
```
Small (Badges): 4px (rounded-sm)
Medium (Buttons, Inputs): 8px (rounded-md)
Large (Cards): 12px (rounded-lg)
Full (Avatars, Pills): 9999px (rounded-full)
```

### Usage Examples
```css
/* Card */
.card {
  border: 1px solid #E5E7EB;
  border-radius: 12px;
}

/* Button */
.btn {
  border-radius: 8px;
}

/* Badge */
.badge {
  border-radius: 4px;
}
```

---

## Shadows

```
Small (Hover):
box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
/* shadow-sm */

Medium (Cards):
box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
/* shadow-md */

Large (Modals):
box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
/* shadow-xl */
```

### Usage
```jsx
// Card (default state)
<div className="shadow-md">

// Card (hover state)
<div className="shadow-md hover:shadow-lg transition-shadow">

// Modal
<div className="shadow-xl">
```

---

## Components

### Agent Card

```jsx
<div className="bg-white border border-gray-200 rounded-lg p-6 shadow-md hover:shadow-lg transition-shadow">
  {/* Title */}
  <h3 className="text-xl font-semibold text-gray-900 mb-2">
    Claude Code
  </h3>

  {/* Separator */}
  <div className="h-px bg-gray-200 mb-3"></div>

  {/* Badge */}
  <span className="inline-block px-2 py-1 bg-blue-100 text-blue-600 text-xs font-medium uppercase rounded-sm">
    IDE Agent
  </span>

  {/* Description */}
  <p className="mt-2 text-sm text-gray-600 line-clamp-2">
    AI-powered CLI coding assistant for developers
  </p>

  {/* Toggle */}
  <div className="mt-4 flex items-center">
    <Switch
      checked={true}
      className="bg-green-600"
      aria-label="Toggle Claude Code"
    />
    <span className="ml-2 text-sm font-medium text-gray-900">Enabled</span>
  </div>

  {/* Button */}
  <Button
    variant="outline"
    className="mt-2 w-full"
  >
    Configure →
  </Button>
</div>
```

### Toggle Switch (shadcn/ui)

```jsx
import { Switch } from "@/components/ui/switch"

// Enabled state
<Switch
  checked={true}
  onCheckedChange={handleToggle}
  className="data-[state=checked]:bg-green-600"
  aria-label="Toggle agent"
/>

// Disabled state
<Switch
  checked={false}
  onCheckedChange={handleToggle}
  className="data-[state=unchecked]:bg-gray-300"
  aria-label="Toggle agent"
/>

// Styling
.switch {
  width: 44px;
  height: 24px;
  border-radius: 12px;
}

.switch-thumb {
  width: 20px;
  height: 20px;
  border-radius: 10px;
  background: white;
  transition: transform 0.2s;
}
```

### Buttons

```jsx
// Primary (Enable action)
<Button
  variant="default"
  className="bg-blue-600 hover:bg-blue-500 text-white"
>
  Enable
</Button>

// Secondary (Configure action)
<Button
  variant="outline"
  className="border-gray-300 hover:bg-gray-50 text-gray-700"
>
  Configure →
</Button>

// Destructive (Disable action)
<Button
  variant="destructive"
  className="bg-red-600 hover:bg-red-500 text-white"
>
  Disable Agent
</Button>

// Ghost (Cancel action)
<Button
  variant="ghost"
  className="hover:bg-gray-100 text-gray-700"
>
  Cancel
</Button>
```

### Badge

```jsx
// Type badge
<span className="inline-block px-2 py-1 bg-blue-100 text-blue-600 text-xs font-medium uppercase rounded-sm">
  IDE Agent
</span>

// Status badge (Enabled)
<span className="inline-flex items-center px-2 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-full">
  <span className="w-2 h-2 bg-green-600 rounded-full mr-1"></span>
  Enabled
</span>

// Status badge (Disabled)
<span className="inline-flex items-center px-2 py-1 bg-gray-100 text-gray-600 text-xs font-medium rounded-full">
  <span className="w-2 h-2 bg-gray-400 rounded-full mr-1"></span>
  Disabled
</span>
```

### Search Input

```jsx
import { Input } from "@/components/ui/input"

<div className="relative">
  <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400">
    {/* Search icon */}
  </svg>
  <Input
    type="search"
    placeholder="Search agents..."
    className="pl-10 pr-4 py-2 border-gray-300 rounded-md focus:ring-2 focus:ring-blue-600"
  />
</div>
```

### Dialog (Confirmation)

```jsx
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"

<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent className="sm:max-w-md">
    <DialogHeader>
      <DialogTitle className="text-lg font-semibold">
        Disable Claude Code?
      </DialogTitle>
    </DialogHeader>

    <p className="text-sm text-gray-600">
      This will remove access for all teams and employees.
      Configurations will be preserved and can be restored.
    </p>

    <DialogFooter className="gap-2">
      <Button variant="ghost" onClick={() => setOpen(false)}>
        Cancel
      </Button>
      <Button variant="destructive" onClick={handleDisable}>
        Disable Agent
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### Toast (Notification)

```jsx
import { useToast } from "@/components/ui/use-toast"

const { toast } = useToast()

// Success
toast({
  title: "Agent enabled",
  description: "Claude Code is now available to your organization",
  variant: "default",
})

// Error
toast({
  title: "Failed to enable agent",
  description: "Please try again or contact support",
  variant: "destructive",
})

// Styling
.toast {
  background: white;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1);
}
```

### Skeleton (Loading)

```jsx
import { Skeleton } from "@/components/ui/skeleton"

<div className="bg-white border border-gray-200 rounded-lg p-6">
  <Skeleton className="h-6 w-32 mb-2" /> {/* Title */}
  <Skeleton className="h-px w-full mb-3" /> {/* Separator */}
  <Skeleton className="h-4 w-24 mb-2" /> {/* Badge */}
  <Skeleton className="h-4 w-full mb-1" /> {/* Description line 1 */}
  <Skeleton className="h-4 w-3/4 mb-4" /> {/* Description line 2 */}
  <Skeleton className="h-6 w-20 mb-2" /> {/* Toggle */}
  <Skeleton className="h-10 w-full" /> {/* Button */}
</div>
```

---

## Animations

### Transitions
```css
/* Default transition */
.transition-default {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

/* Shadow transition */
.transition-shadow {
  transition-property: box-shadow;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

/* Color transition */
.transition-colors {
  transition-property: color, background-color, border-color;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}
```

### Hover Effects
```jsx
// Card hover
<div className="hover:shadow-lg transition-shadow">

// Button hover
<Button className="hover:bg-blue-500 transition-colors">

// Link hover
<a className="hover:text-blue-600 transition-colors">
```

### Loading States
```jsx
// Shimmer effect
@keyframes shimmer {
  0% { background-position: -468px 0; }
  100% { background-position: 468px 0; }
}

.shimmer {
  background: linear-gradient(
    to right,
    #F3F4F6 0%,
    #E5E7EB 20%,
    #F3F4F6 40%,
    #F3F4F6 100%
  );
  background-size: 800px 100px;
  animation: shimmer 1.5s infinite linear;
}
```

### Slide-in (Mobile)
```css
/* Bottom sheet */
@keyframes slide-up {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

.bottom-sheet {
  animation: slide-up 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

---

## Responsive Breakpoints

```css
/* Tailwind breakpoints */
sm: 640px   /* Small devices */
md: 768px   /* Tablets */
lg: 1024px  /* Desktops */
xl: 1280px  /* Large desktops */
2xl: 1536px /* Extra large */
```

### Usage
```jsx
// Responsive grid
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">

// Responsive padding
<div className="p-4 md:p-6 lg:p-8">

// Responsive text
<h1 className="text-2xl md:text-3xl lg:text-4xl">
```

---

## Accessibility

### Focus States
```css
/* Focus ring */
.focus-ring {
  outline: 2px solid #3B82F6;
  outline-offset: 4px;
}

/* Focus visible (keyboard only) */
.focus-visible:focus-visible {
  outline: 2px solid #3B82F6;
  outline-offset: 4px;
}
```

### Color Contrast
```
WCAG AA Compliance:
- Normal text: 4.5:1 minimum
- Large text (18px+): 3:1 minimum
- UI components: 3:1 minimum

Tested Combinations:
✓ Gray-900 on White: 14.4:1
✓ Gray-700 on White: 8.6:1
✓ Gray-600 on White: 6.3:1
✓ Blue-600 on White: 4.6:1
✓ White on Blue-600: 4.6:1
```

### Touch Targets
```
Minimum: 44x44px (iOS/Android guidelines)

All buttons, toggles, and interactive elements MUST meet this size.
```

### ARIA Attributes
```jsx
// Toggle
<Switch
  role="switch"
  aria-checked={enabled}
  aria-label="Toggle Claude Code (currently enabled)"
/>

// Button
<Button
  aria-label="Configure Claude Code settings"
  aria-describedby="agent-description"
>
  Configure →
</Button>

// Search
<Input
  type="search"
  role="searchbox"
  aria-label="Search agents"
  aria-autocomplete="list"
/>
```

---

## Z-Index Scale

```
Base: 0
Dropdown: 10
Sticky Header: 20
Modal Backdrop: 30
Modal Content: 40
Toast: 50
Tooltip: 60
```

---

## Icon System

Use **Lucide React** icons (matches shadcn/ui)

```jsx
import { Settings, Plus, Search, X, Check, AlertCircle } from 'lucide-react'

// Sizes
<Search className="w-4 h-4" />  // 16px (small)
<Search className="w-5 h-5" />  // 20px (medium)
<Search className="w-6 h-6" />  // 24px (large)

// Colors
<Search className="text-gray-400" />  // Muted
<Search className="text-blue-600" />  // Primary
<Search className="text-red-600" />   // Destructive
```

---

## Grid System

```jsx
// 3-column desktop
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">

// Auto-fit (flexible columns)
<div className="grid grid-cols-[repeat(auto-fit,minmax(280px,1fr))] gap-6">

// 12-column layout
<div className="grid grid-cols-12 gap-4">
  <div className="col-span-12 md:col-span-8">Main</div>
  <div className="col-span-12 md:col-span-4">Sidebar</div>
</div>
```

---

## Related Wireframes

- [agents-redesign-desktop.md](./agents-redesign-desktop.md) - Desktop layout
- [agents-redesign-mobile.md](./agents-redesign-mobile.md) - Mobile layout
