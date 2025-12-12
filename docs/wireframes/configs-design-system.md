# Agent Configurations - Design System Reference

## Color Palette

### Primary Colors

```
Primary Blue
#3B82F6  rgb(59, 130, 246)
Usage: Primary actions, links, focus states

Primary Blue Hover
#2563EB  rgb(37, 99, 235)
Usage: Hover states for primary elements

Primary Blue Light
#EFF6FF  rgb(239, 246, 255)
Usage: Selected backgrounds, highlights
```

### Semantic Colors

```
Success Green
#10B981  rgb(16, 185, 129)
Usage: Enabled status, success messages

Success Green Light
#D1FAE5  rgb(209, 250, 229)
Usage: Success badge backgrounds

Error Red
#EF4444  rgb(239, 68, 68)
Usage: Error text, destructive actions

Error Red Light
#FEE2E2  rgb(254, 226, 226)
Usage: Error banner backgrounds

Warning Amber
#F59E0B  rgb(245, 158, 11)
Usage: Warning states, caution messages

Warning Amber Light
#FEF3C7  rgb(254, 243, 199)
Usage: Warning banner backgrounds
```

### Neutral Grays

```
Gray 50 - #F9FAFB   Backgrounds
Gray 100 - #F3F4F6  Hover states, disabled
Gray 200 - #E5E7EB  Borders, dividers
Gray 300 - #D1D5DB  Input borders
Gray 400 - #9CA3AF  Icons, placeholders
Gray 500 - #6B7280  Secondary text
Gray 600 - #4B5563  Body text
Gray 700 - #374151  Labels
Gray 800 - #1F2937  Headings
Gray 900 - #111827  Primary text
```

## Typography

### Font Families

```css
/* Primary Font - UI Text */
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;

/* Monospace Font - Code/JSON */
font-family: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
```

### Font Sizes & Weights

```css
/* Headings */
h1: 36px / 2.25rem - Semibold 600
h2: 30px / 1.875rem - Semibold 600
h3: 24px / 1.5rem - Semibold 600
h4: 20px / 1.25rem - Semibold 600
h5: 18px / 1.125rem - Medium 500
h6: 16px / 1rem - Medium 500

/* Body Text */
Large: 18px / 1.125rem - Regular 400
Base: 16px / 1rem - Regular 400
Small: 14px / 0.875rem - Regular 400
XSmall: 12px / 0.75rem - Regular 400

/* Labels */
Label: 14px / 0.875rem - Medium 500
Label Small: 12px / 0.75rem - Medium 500

/* Buttons */
Button Large: 16px / 1rem - Medium 500
Button Base: 14px / 0.875rem - Medium 500
Button Small: 13px / 0.8125rem - Medium 500

/* Code */
Code: 13px / 0.8125rem - Regular 400 (Mono)
Code Small: 12px / 0.75rem - Regular 400 (Mono)
```

### Line Heights

```css
Tight: 1.25
Normal: 1.5
Relaxed: 1.625
Loose: 2
```

## Spacing System

### Base Unit: 4px

```
0   = 0px
1   = 4px
2   = 8px
3   = 12px
4   = 16px
5   = 20px
6   = 24px
7   = 28px
8   = 32px
10  = 40px
12  = 48px
16  = 64px
20  = 80px
24  = 96px
```

### Common Usage

```css
/* Component Padding */
Button: 12px 24px (3 6)
Input: 12px 16px (3 4)
Card: 16px (4) or 24px (6)
Modal: 32px (8)
Page: 16px (4) mobile, 24px (6) desktop

/* Component Gaps */
Form Fields: 16px (4)
Filter Bar: 12px (3)
Table Cells: 16px (4)
Action Menu Items: 8px (2)

/* Section Spacing */
Between Sections: 32px (8)
Between Subsections: 24px (6)
Between Elements: 16px (4)
```

## Border Radius

```css
None: 0px
Small: 4px   (buttons, badges)
Base: 6px    (inputs, cards)
Medium: 8px  (modals, dropdowns)
Large: 12px  (containers)
Full: 9999px (pills, badges)
```

## Shadows

```css
/* Elevation Levels */
Small: 0 1px 3px rgba(0, 0, 0, 0.05)
Usage: Cards, table rows on hover

Base: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)
Usage: Dropdowns, popovers

Medium: 0 4px 6px rgba(0, 0, 0, 0.1)
Usage: Floating elements

Large: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)
Usage: Modals

XLarge: 0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04)
Usage: Overlays

/* Focus Shadow */
Focus: 0 0 0 3px rgba(59, 130, 246, 0.5)
```

## Component Specifications

### Buttons

```css
/* Primary Button */
height: 44px
padding: 12px 24px
background: #3B82F6
color: #FFFFFF
border-radius: 6px
font: Inter Medium 14px
hover: #2563EB
focus: ring-2 ring-blue-500 ring-offset-2
disabled: bg-gray-300, not-allowed

/* Secondary Button */
height: 44px
padding: 12px 24px
background: #FFFFFF
border: 1px solid #D1D5DB
color: #374151
border-radius: 6px
font: Inter Medium 14px
hover: bg-gray-50
focus: ring-2 ring-blue-500 ring-offset-2

/* Tertiary Button */
height: 44px
padding: 12px 24px
background: transparent
color: #6B7280
font: Inter Medium 14px
hover: bg-gray-100
focus: ring-2 ring-blue-500 ring-offset-2

/* Icon Button */
width: 32px
height: 32px
padding: 0
background: transparent
color: #9CA3AF
hover: bg-gray-100, color #4B5563
touch-target: 44x44px (with invisible padding)
```

### Form Controls

```css
/* Text Input */
height: 44px
padding: 12px 16px
border: 1px solid #D1D5DB
border-radius: 6px
font: Inter Regular 14px
placeholder: #9CA3AF
focus: border-2 solid #3B82F6
error: border-2 solid #EF4444

/* Select Dropdown */
height: 44px
padding: 12px 16px 12px 16px
border: 1px solid #D1D5DB
border-radius: 6px
font: Inter Regular 14px
chevron: right 12px, gray-400
focus: border-2 solid #3B82F6

/* Checkbox */
width: 20px
height: 20px
border: 2px solid #D1D5DB
border-radius: 4px
checked: bg-blue-500, white checkmark
focus: ring-2 ring-blue-500 ring-offset-2

/* Radio Button */
width: 20px
height: 20px
border: 2px solid #D1D5DB
border-radius: 50%
checked: bg-blue-500, white dot (8px)
focus: ring-2 ring-blue-500 ring-offset-2

/* Toggle Switch */
width: 48px
height: 28px
background: #D1D5DB (off), #3B82F6 (on)
border-radius: 14px
knob: 24px circle, white, shadow-sm
transition: 150ms ease
focus: ring-2 ring-blue-500 ring-offset-2

/* Range Slider */
track-height: 4px
track-color: #E5E7EB
fill-color: #3B82F6
thumb-size: 16px
thumb-color: #3B82F6
thumb-shadow: 0 1px 3px rgba(0,0,0,0.1)
```

### Badges

```css
/* Status Badge - Enabled */
height: 28px
padding: 4px 12px
background: #D1FAE5
color: #065F46
border-radius: 12px
font: Inter Medium 12px

/* Status Badge - Disabled */
height: 28px
padding: 4px 12px
background: #F3F4F6
color: #6B7280
border-radius: 12px
font: Inter Medium 12px

/* Count Badge */
min-width: 20px
height: 20px
padding: 2px 6px
background: #EF4444
color: #FFFFFF
border-radius: 10px
font: Inter Bold 11px
```

### Cards

```css
/* Table Card (Desktop) */
min-height: 88px
padding: 16px 24px
background: #FFFFFF
border: 1px solid #E5E7EB
border-radius: 0 (table rows)
hover: bg-gray-50
shadow: none (default), shadow-sm (hover)

/* Config Card (Mobile) */
min-height: 160px
padding: 16px
background: #FFFFFF
border: 1px solid #E5E7EB
border-radius: 8px
shadow: 0 1px 3px rgba(0,0,0,0.05)
tap: scale-98, shadow-md
```

### Modals

```css
/* Modal Overlay */
background: rgba(0, 0, 0, 0.5)
backdrop-filter: blur(2px)

/* Modal Container */
width: 640px (desktop), calc(100vw - 32px) (mobile)
max-height: 90vh
background: #FFFFFF
border-radius: 12px
padding: 32px
shadow: 0 20px 25px rgba(0,0,0,0.1)

/* Modal Header */
font: Inter Semibold 20px
color: #111827
border-bottom: 1px solid #E5E7EB
padding-bottom: 16px
margin-bottom: 24px

/* Modal Close Button */
width: 24px
height: 24px
color: #9CA3AF
hover: bg-gray-100, color-gray-600
position: absolute, top-right 16px
```

### Dropdowns

```css
/* Dropdown Container */
min-width: 160px
max-width: 320px
background: #FFFFFF
border: 1px solid #E5E7EB
border-radius: 8px
shadow: 0 10px 15px rgba(0,0,0,0.1)
padding: 8px 0

/* Dropdown Item */
height: 44px (touch-friendly)
padding: 12px 16px
font: Inter Regular 14px
color: #374151
hover: bg-gray-50
active: bg-blue-50, color-blue-600
focus: bg-gray-100

/* Dropdown Divider */
height: 1px
background: #E5E7EB
margin: 4px 0

/* Dropdown Icon */
width: 20px
height: 20px
color: #9CA3AF
margin-right: 12px
```

### Tables

```css
/* Table Container */
width: 100%
background: #FFFFFF
border: 1px solid #E5E7EB
border-radius: 8px

/* Table Header */
height: 48px
background: #F9FAFB
border-bottom: 1px solid #E5E7EB
padding: 12px 24px
font: Inter Medium 13px
color: #6B7280
text-transform: uppercase
letter-spacing: 0.05em

/* Table Row */
min-height: 88px
border-bottom: 1px solid #E5E7EB
padding: 16px 24px
hover: bg-gray-50

/* Table Cell */
padding: 16px 24px
vertical-align: middle
font: Inter Regular 14px
color: #374151
```

### Toasts

```css
/* Toast Container */
min-width: 320px
max-width: 480px
padding: 16px
background: #FFFFFF
border: 1px solid #E5E7EB
border-radius: 8px
shadow: 0 10px 15px rgba(0,0,0,0.1)

/* Success Toast */
border-left: 4px solid #10B981
icon: checkmark-circle, green-500

/* Error Toast */
border-left: 4px solid #EF4444
icon: x-circle, red-500

/* Warning Toast */
border-left: 4px solid #F59E0B
icon: exclamation-triangle, amber-500

/* Info Toast */
border-left: 4px solid #3B82F6
icon: information-circle, blue-500
```

### Skeletons

```css
/* Skeleton Base */
background: linear-gradient(
  90deg,
  #F3F4F6 0%,
  #E5E7EB 50%,
  #F3F4F6 100%
)
background-size: 200% 100%
animation: shimmer 1.5s ease-in-out infinite

/* Skeleton Text Line */
height: 16px (small), 20px (base), 24px (large)
border-radius: 4px

/* Skeleton Button */
height: 44px
width: 120px
border-radius: 6px

/* Skeleton Avatar */
width: 40px
height: 40px
border-radius: 50%
```

## Transitions & Animations

```css
/* Standard Transition */
transition: all 150ms ease-in-out

/* Hover Effects */
transition: background-color 150ms, color 150ms, transform 100ms

/* Modal Enter/Exit */
enter: opacity 200ms, scale 200ms ease-out
exit: opacity 150ms, scale 150ms ease-in

/* Dropdown Enter/Exit */
enter: opacity 100ms, translateY 100ms ease-out
exit: opacity 75ms ease-in

/* Toast Enter/Exit */
enter: translateX 300ms spring(300, 30, 0, 0)
exit: opacity 200ms, scale 200ms ease-in

/* Skeleton Shimmer */
@keyframes shimmer {
  0% { background-position: 200% 0 }
  100% { background-position: -200% 0 }
}
```

## Responsive Breakpoints

```css
/* Mobile First Approach */
xs: 0px      (min-width)
sm: 640px    (min-width)
md: 768px    (min-width)
lg: 1024px   (min-width)
xl: 1280px   (min-width)
2xl: 1536px  (min-width)

/* Common Usage */
Mobile: 320px - 767px
Tablet: 768px - 1023px
Desktop: 1024px+
```

## Icons

### Icon Library: Heroicons

```css
/* Icon Sizes */
xs: 16x16px
sm: 20x20px
base: 24x24px
lg: 28x28px
xl: 32x32px

/* Icon Weights */
outline: 1.5px stroke (default)
solid: filled (emphasis)

/* Common Icons Used */
Magnifying Glass (search)
Chevron Down (dropdowns)
Ellipsis Vertical (actions menu)
Plus (create)
Pencil (edit)
Duplicate (copy)
Toggle (enable/disable)
Trash (delete)
X Mark (close)
Check (success)
Exclamation Triangle (warning)
Information Circle (info)
Cog (settings/config)
```

## Accessibility

### Focus Indicators

```css
/* Keyboard Focus */
outline: 2px solid #3B82F6
outline-offset: 2px
border-radius: inherit

/* Alternative Focus (inset) */
box-shadow: inset 0 0 0 2px #3B82F6

/* Skip Link */
position: absolute
left: -9999px
focus: left-4, top-4, bg-blue-500, color-white
```

### Color Contrast Ratios

```
WCAG AA (Minimum):
- Normal text: 4.5:1
- Large text (18px+): 3:1
- UI components: 3:1

WCAG AAA (Enhanced):
- Normal text: 7:1
- Large text: 4.5:1

Our Compliance:
✓ Gray 900 on White: 16.5:1 (AAA)
✓ Gray 700 on White: 10.7:1 (AAA)
✓ Gray 600 on White: 7.8:1 (AAA)
✓ Gray 500 on White: 4.6:1 (AA)
✓ Blue 600 on White: 5.9:1 (AA)
✓ Green 600 on Light Green: 4.8:1 (AA)
✓ Red 700 on Light Red: 5.2:1 (AA)
```

### Screen Reader Only

```css
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
```

## Z-Index Scale

```css
/* Layering Order */
base: 0           (default content)
dropdown: 10      (dropdowns, popovers)
sticky: 20        (sticky headers)
modal-backdrop: 30 (modal overlays)
modal: 40         (modal content)
popover: 50       (tooltips, hints)
toast: 60         (notifications)
```

## Print Styles

```css
@media print {
  /* Hide UI Chrome */
  .no-print { display: none }

  /* Optimize Table */
  table { page-break-inside: avoid }
  tr { page-break-inside: avoid }

  /* Black & White Friendly */
  * {
    color: black !important;
    background: white !important;
  }

  /* Show URLs in Links */
  a[href]:after {
    content: " (" attr(href) ")";
  }
}
```

## Dark Mode Support (Future)

```css
@media (prefers-color-scheme: dark) {
  /* Invert Palette */
  --bg-primary: #111827
  --bg-secondary: #1F2937
  --text-primary: #F9FAFB
  --text-secondary: #D1D5DB
  --border: #374151

  /* Adjust Shadows */
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.5)
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.5)

  /* Semantic Colors */
  --success: #34D399 (lighter green)
  --error: #F87171 (lighter red)
  --warning: #FBBF24 (lighter amber)
}
```

## Component Library Mapping

### shadcn/ui Components Used

| Component | Config Page Usage |
|-----------|-------------------|
| Table | Main config list |
| Button | All CTAs |
| Dialog | Create/edit modal |
| Dropdown Menu | Filters, actions |
| Input | Search, form fields |
| Select | Agent, team, employee selection |
| Badge | Status indicators |
| Skeleton | Loading state |
| Alert | Error banners |
| Toast | Notifications |
| Radio Group | Assignment type |
| Checkbox | Feature toggles |
| Slider | Temperature range |
| Switch | Enable/disable toggle |
| Card | Mobile config cards |
| Separator | Dividers |

### Custom Components Needed

| Component | Purpose |
|-----------|---------|
| ConfigTable | Specialized table with preview |
| ConfigCard | Mobile card layout |
| ConfigPreview | 3-line truncated config display |
| FilterBar | Multi-filter toolbar |
| CreateWizard | 2-step modal flow |
| JSONEditor | Syntax-highlighted editor |
| EmptyState | Contextual empty states |

## Implementation Notes

### CSS Framework: Tailwind CSS

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        // Custom brand colors
      },
      spacing: {
        // Custom spacing values
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
```

### CSS Variables

```css
:root {
  /* Colors */
  --color-primary: 59 130 246;
  --color-success: 16 185 129;
  --color-error: 239 68 68;

  /* Spacing */
  --spacing-unit: 4px;

  /* Timing */
  --transition-fast: 150ms;
  --transition-base: 200ms;
  --transition-slow: 300ms;

  /* Z-Index */
  --z-dropdown: 10;
  --z-modal: 40;
  --z-toast: 60;
}
```

---

**Design System Version:** 1.0
**Last Updated:** 2025-01-15
**Maintained By:** Product Designer
**Status:** ✅ Approved for Implementation
