# PartFlow Frontend Design System
## القانون البصري للمشروع

---

## نظرة عامة

هذه الوثيقة تحدد **اللغة البصرية الموحدة** لـ PartFlow. الهدف هو جعل كل المكونات والصفحات تتحدث "لغة واحدة" بصرياً، مما يعطي المستخدم إحساس أنه داخل نظام احترافي موحد.

---

## 1. Color System

### Primary Colors
- **Primary Blue**: `#3B82F6` (دون تغيير)
- **Primary Dark**: `#1E40AF`
- **Primary Light**: `#60A5FA`

### Secondary Colors
- **Success Green**: `#10B981`
- **Warning Yellow**: `#F59E0B`
- **Error Red**: `#EF4444`
- **Info Cyan**: `#06B6D4`

### Neutral Colors
- **Background Primary**: `#FFFFFF`
- **Background Secondary**: `#F3F4F6`
- **Background Tertiary**: `#E5E7EB`
- **Text Primary**: `#111827`
- **Text Secondary**: `#6B7280`
- **Text Tertiary**: `#9CA3AF`

### Semantic Colors
- **Brand**: `#3B82F6`
- **Accent**: `#8B5CF6`
- **Warning**: `#F59E0B`
- **Danger**: `#EF4444`
- **Success**: `#10B981`

---

## 2. Typography

### Font Families
- **Primary**: "Tajawal", "Cairo", sans-serif (للعربية)
- **Secondary**: "Inter", "Segoe UI", sans-serif (للإنجليزية)

### Font Sizes
- **Display Large**: `32px` / `2rem`
- **Display Medium**: `24px` / `1.5rem`
- **Display Small**: `20px` / `1.25rem`
- **Heading Large**: `18px` / `1.125rem`
- **Heading Medium**: `16px` / `1rem`
- **Body Large**: `16px` / `1rem`
- **Body Medium**: `14px` / `0.875rem`
- **Body Small**: `12px` / `0.75rem`
- **Caption**: `11px` / `0.6875rem`

### Font Weights
- **Bold**: `700`
- **SemiBold**: `600`
- **Medium**: `500`
- **Regular**: `400`
- **Light**: `300`

### Line Heights
- **Tight**: `1.25`
- **Normal**: `1.5`
- **Relaxed**: `1.75`

---

## 3. Spacing System

### Base Unit
- **Base**: `4px`

### Scale
- **XS**: `4px` / `0.25rem`
- **SM**: `8px` / `0.5rem`
- **MD**: `16px` / `1rem`
- **LG**: `24px` / `1.5rem`
- **XL**: `32px` / `2rem`
- **2XL**: `48px` / `3rem`
- **3XL**: `64px` / `4rem`

### Usage Guidelines
- **Component Internal**: `SM` to `MD`
- **Component External**: `MD` to `LG`
- **Section Spacing**: `LG` to `XL`
- **Page Level**: `XL` to `2XL`

---

## 4. Border Radius

### Scale
- **None**: `0px`
- **SM**: `4px`
- **MD**: `8px`
- **LG**: `12px`
- **XL**: `16px`
- **Full**: `9999px`

### Usage Guidelines
- **Buttons**: `MD` to `LG`
- **Cards**: `LG`
- **Inputs**: `MD`
- **Dialogs**: `XL`
- **Badges**: `SM` to `Full`

---

## 5. Shadows

### Elevation System
- **XS**: `0 1px 2px rgba(0, 0, 0, 0.05)`
- **SM**: `0 1px 3px rgba(0, 0, 0, 0.1)`
- **MD**: `0 4px 6px rgba(0, 0, 0, 0.1)`
- **LG**: `0 10px 15px rgba(0, 0, 0, 0.1)`
- **XL**: `0 20px 25px rgba(0, 0, 0, 0.15)`

### Usage Guidelines
- **Cards**: `SM` to `MD`
- **Dialogs**: `LG` to `XL`
- **Dropdowns**: `MD`
- **Tooltips**: `XS` to `SM`

---

## 6. Buttons

### Primary Button
- **Background**: Primary Blue
- **Text**: White
- **Radius**: `MD`
- **Padding**: `MD` vertical, `LG` horizontal
- **Font Weight**: `SemiBold`
- **Hover**: Primary Dark
- **Active**: Primary Dark + slight offset

### Secondary Button
- **Background**: Background Secondary
- **Text**: Text Primary
- **Border**: `1px solid` Background Tertiary
- **Radius**: `MD`
- **Padding**: `MD` vertical, `LG` horizontal
- **Font Weight**: `SemiBold`
- **Hover**: Background Tertiary

### Ghost Button
- **Background**: Transparent
- **Text**: Primary Blue
- **Radius**: `MD`
- **Padding**: `MD` vertical, `LG` horizontal
- **Font Weight**: `SemiBold`
- **Hover**: Background Secondary

### Danger Button
- **Background**: Error Red
- **Text**: White
- **Radius**: `MD`
- **Padding**: `MD` vertical, `LG` horizontal
- **Font Weight**: `SemiBold`
- **Hover**: Darker Red

### Button Sizes
- **SM**: Height `32px`, Font `Body Small`
- **MD**: Height `40px`, Font `Body Medium`
- **LG**: Height `48px`, Font `Body Large`

---

## 7. Cards

### Base Card
- **Background**: White
- **Border**: `1px solid` Background Tertiary
- **Radius**: `LG`
- **Shadow**: `SM`
- **Padding**: `LG`

### Card Variants
- **Elevated**: Shadow `MD`
- **Bordered**: Thicker border
- **Minimal**: No shadow, no border
- **Interactive**: Hover shadow `MD`

### Card Structure
```
┌─────────────────────────┐
│  [Header: Title + Icon] │
├─────────────────────────┤
│                         │
│   [Content Area]        │
│                         │
├─────────────────────────┤
│  [Footer: Actions]      │
└─────────────────────────┘
```

---

## 8. Tables

### Primitive Table (ui/table)
- **Background**: White
- **Header Background**: Background Secondary
- **Border**: `1px solid` Background Tertiary
- **Radius**: `MD`
- **Cell Padding**: `MD` vertical, `LG` horizontal

### Business Table (tables/)
- **Inherits**: Primitive table
- **Additional Features**:
  - Selection
  - Sorting
  - Filtering
  - Pagination
  - Actions column
  - Bulk actions

### Table Visual Hierarchy
- **Header Row**: Bold, darker background
- **Data Rows**: Zebra striping (alternating backgrounds)
- **Hover State**: Background Secondary
- **Selected State**: Primary Blue background, white text

---

## 9. Forms

### Input Fields
- **Background**: White
- **Border**: `1px solid` Background Tertiary
- **Radius**: `MD`
- **Padding**: `MD` vertical, `LG` horizontal
- **Font**: `Body Medium`
- **Focus Border**: Primary Blue, `2px`
- **Error Border**: Error Red

### Input States
- **Default**: Gray border
- **Focus**: Blue border, `2px`
- **Error**: Red border
- **Disabled**: Gray background, readonly
- **Success**: Green border

### Form Layout
- **Label**: `Body Small`, Text Secondary, margin-bottom `SM`
- **Field**: Full width
- **Helper Text**: `Caption`, Text Tertiary, margin-top `SM`
- **Error Text**: `Caption`, Error Red, margin-top `SM`

---

## 10. Dialogs

### Primitive Dialog (ui/dialog)
- **Background**: White
- **Radius**: `XL`
- **Shadow**: `XL`
- **Max Width**: `600px`
- **Padding**: `XL`

### Business Dialog (dialog/)
- **Inherits**: Primitive dialog
- **Additional Features**:
  - Title bar
  - Action buttons
  - Close button
  - Backdrop

### Dialog Structure
```
┌─────────────────────────┐
│  [Title]          [×]   │
├─────────────────────────┤
│                         │
│   [Content]             │
│                         │
├─────────────────────────┤
│  [Cancel]    [Confirm]  │
└─────────────────────────┘
```

---

## 11. Badges

### Badge Variants
- **Default**: Background Secondary, Text Primary
- **Primary**: Primary Blue background, White text
- **Success**: Success Green background, White text
- **Warning**: Warning Yellow background, White text
- **Error**: Error Red background, White text

### Badge Sizes
- **SM**: Height `20px`, Font `Caption`
- **MD**: Height `24px`, Font `Body Small`
- **LG**: Height `28px`, Font `Body Medium`

### Badge Radius
- **SM**: `SM`
- **MD**: `MD`
- **LG**: `Full`

---

## 12. Navigation

### Sidebar
- **Width**: `256px`
- **Background**: Dark (#1F2937)
- **Text**: White
- **Active Item**: Primary Blue background
- **Hover Item**: Slightly lighter background
- **Item Height**: `48px`
- **Item Padding**: `MD` horizontal

### Topbar
- **Height**: `64px`
- **Background**: White
- **Border Bottom**: `1px solid` Background Tertiary
- **Padding**: `LG` horizontal

### Breadcrumbs
- **Font**: `Body Small`
- **Color**: Text Secondary
- **Separator**: `/`
- **Active**: Text Primary, Bold

---

## 13. Page Layout

### Standard Page Structure
```
┌─────────────────────────────────────┐
│          [Topbar]                   │
├──────────┬──────────────────────────┤
│          │  [Page Header]           │
│          │  [Title + Actions]       │
│ [Sidebar]├──────────────────────────┤
│          │  [Content]               │
│          │  [Stats/Filters]         │
│          │  [Main Data]             │
│          │  [Secondary Info]        │
└──────────┴──────────────────────────┘
```

### Content Width
- **Maximum**: `1280px`
- **Default**: `100%`
- **Narrow**: `800px` (for forms/settings)

### Section Spacing
- **Between Sections**: `XL`
- **Within Sections**: `LG`

---

## 14. Responsive Design

### Breakpoints
- **Mobile**: `< 640px`
- **Tablet**: `640px - 1024px`
- **Desktop**: `> 1024px`
- **Large Desktop**: `> 1280px`

### Responsive Behaviors
- **Mobile**: Sidebar becomes drawer, tables become cards
- **Tablet**: Sidebar collapses, grids reduce columns
- **Desktop**: Full layout

---

## 15. RTL Support

### Direction
- **Default**: RTL (right-to-left)
- **Components**: Must support RTL natively
- **Spacing**: Logical properties (margin-inline-start instead of margin-left)
- **Icons**: Must flip for RTL

### Typography
- **Font**: Tajawal/Cairo for Arabic
- **Alignment**: Right-aligned by default
- **Numbers**: Can use Arabic-Indic digits optionally

---

## 16. Dark Mode

### Color Mapping
- **Background Primary**: Dark Gray (#1F2937)
- **Background Secondary**: Darker Gray (#111827)
- **Text Primary**: Light Gray (#F9FAFB)
- **Text Secondary**: Medium Gray (#D1D5DB)
- **Borders**: Darker Gray (#374151)

### Implementation
- **Toggle**: In settings
- **Persistence**: localStorage
- **Transition**: Smooth color transition

---

## 17. Component Hierarchy

### Level 1: Primitives (components/ui/)
```
Button, Input, Card, Table, Dialog, Badge, etc.
```

### Level 2: Shared Components (components/)
```
Navigation, Forms, Tables, Dialog, Theme, etc.
```

### Level 3: Feature Components (features/*/components/)
```
Inventory Components, Sales Components, etc.
```

### Level 4: Page Components (features/*/pages/)
```
Full pages composed of above levels
```

---

## 18. Page Composition Rules

### Standard Page
1. **Page Header**: Title + Actions
2. **Summary Stats**: 3-4 key metrics
3. **Search & Filters**: Toolbar
4. **Main Content**: Data grid/table
5. **Secondary Info**: Details, related data

### Visual Hierarchy
- **Primary**: Large, bold, primary colors
- **Secondary**: Medium, regular text
- **Tertiary**: Small, muted text

### Content Priority
```
PRIMARY: Sales, Inventory, Customers
SECONDARY: Purchases, Suppliers, Debts
MANAGEMENT: Reports, Expenses, Audit
SYSTEM: Settings, Notifications, Shortcuts
```

---

## 19. Animation & Transitions

### Transition Speed
- **Fast**: `150ms`
- **Normal**: `200ms`
- **Slow**: `300ms`

### Easing
- **Ease In**: `cubic-bezier(0.4, 0, 1, 1)`
- **Ease Out**: `cubic-bezier(0, 0, 0.2, 1)`
- **Ease In Out**: `cubic-bezier(0.4, 0, 0.2, 1)`

### Use Cases
- **Hover**: Fast, ease out
- **Modal**: Normal, ease in out
- **Page Load**: Slow, ease out

---

## 20. Accessibility

### Color Contrast
- **WCAG AA**: 4.5:1 for normal text
- **WCAG AAA**: 7:1 for large text

### Keyboard Navigation
- **Tab Order**: Logical
- **Focus States**: Visible outline
- **Shortcuts**: Documented

### Screen Readers
- **ARIA Labels**: On interactive elements
- **Semantic HTML**: Proper heading hierarchy
- **Alt Text**: On images

---

## 21. Iconography

### Icon System
- **Library**: Lucide React / Heroicons
- **Size**: `16px`, `20px`, `24px`, `32px`
- **Color**: Inherit from text color
- **Stroke**: `2px` (default)

### Icon Usage
- **Primary Actions**: Larger, more prominent
- **Secondary Actions**: Smaller, less prominent
- **Decorative**: Minimal, subtle

---

## 22. Data Visualization

### Charts
- **Colors**: Primary palette
- **Grid**: Light gray lines
- **Labels**: Small, muted text
- **Tooltips**: Dark background, white text

### Progress Bars
- **Background**: Background Secondary
- **Fill**: Primary Blue (or semantic color)
- **Height**: `8px` to `12px`
- **Radius**: `Full`

---

## 23. Status & Feedback

### Loading States
- **Skeleton**: Gray placeholders
- **Spinner**: Primary blue animation
- **Progress**: Percentage indicator

### Empty States
- **Icon**: Large, muted
- **Message**: Clear, helpful
- **Action**: Primary button

### Error States
- **Color**: Error red
- **Message**: Clear, actionable
- **Recovery**: Next steps

---

## 24. Component Audit Checklist

### Consistency Check
- [ ] Colors match system
- [ ] Typography matches system
- [ ] Spacing matches system
- [ ] Border radius matches system
- [ ] Shadows match system
- [ ] RTL support
- [ ] Dark mode support
- [ ] Responsive behavior

### Quality Check
- [ ] Accessibility (keyboard, screen reader)
- [ ] Performance (lazy loading, code splitting)
- [ ] Error handling
- [ ] Loading states
- [ ] Empty states

---

## 25. Implementation Priority

### Phase 1: Foundation
1. Color system implementation
2. Typography system
3. Spacing system
4. Border radius system

### Phase 2: Components
1. Button consistency
2. Card consistency
3. Input consistency
4. Table consistency

### Phase 3: Features
1. Dashboard page composition
2. Inventory page composition
3. Sales page composition
4. Customers page composition

### Phase 4: Polish
1. Visual hierarchy refinement
2. Animation polish
3. Accessibility improvements
4. Performance optimization

---

## 26. Maintenance Guidelines

### Adding New Components
1. Check existing system first
2. Follow established patterns
3. Update this document
4. Add to component library

### Modifying Existing Components
1. Consider impact on other features
2. Maintain backward compatibility
3. Update documentation
4. Test across features

### Regular Audits
- **Monthly**: Component consistency check
- **Quarterly**: Design system review
- **Annually**: Full system refresh

---

## الخلاصة

هذا Design System هو **القانون البصري** لـ PartFlow. كل مكون جديد أو صفحة جديدة يجب أن تتبع هذه القواعد لضمان اتساق النظام البصري.

**الهدف**: جعل كل المكونات والصفحات تتحدث "لغة واحدة" بصرياً، مما يعطي المستخدم إحساس أنه داخل نظام احترافي موحد.

---

## المراجع

- Fynexa Design System (للمقارنة والمبدأ فقط)
- Tailwind CSS Defaults
- Material Design Guidelines
- Apple Human Interface Guidelines
- Web Content Accessibility Guidelines (WCAG)
