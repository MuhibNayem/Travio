# Travio Frontend Engineering & Design Rule Book

> **Status**: Living Document  
> **Version**: 1.0.0  
> **Objective**: Building a FAANG-scale, high-engagement Transportation & Ticketing Platform.

---

## 1. Design Philosophy: "Liquid Glass" Aesthetic

Our design language is inspired by Apple's visionOS and modern "Liquid Glass" aesthetics. It combines high translucency, vibrant background mesh gradients, and physically grounded motion.

### 1.1 Core Visual Pillars
1.  **Materiality (Glass)**:
    *   UI elements are not solid blocks; they are panes of glass floating above a vibrant background.
    *   **Token**: `bg-white/60 dark:bg-black/40 backdrop-blur-xl border border-white/20 shadow-lg`.
    *   **Rule**: Never use solid white (`#ffffff`) for main containers. Always use semi-transparent backgrounds with `backdrop-blur`.

2.  **Vibrancy (Color)**:
    *   Use **OKLCH** color space for perceptual uniformity and vibrant gradients.
    *   Backgrounds should generally be substrate layers of moving color (mesh gradients) that show *through* the glass active elements.
    *   **Primary Action**: Electric Blue / Vivid Indigo.
    *   **Success/Status**: soothing, not harsh (e.g., Teal instead of pure Green).

3.  **Depth & Elevation**:
    *   Use layered shadows (ambient + direct) to create depth.
    *   **Rule**: "Lift" active elements on hover.
    *   **Z-Index Hierarchy**: Background -> Mesh Gradient -> Content Layer 1 (Glass) -> Content Layer 2 (Float).

4.  **Motion Design**:
    *   Everything must feel "alive".
    *   **Micro-interactions**: Buttons scale down slightly on click (`scale-95`).
    *   **Transitions**: Smooth layout shifts using `svelte/transition` (fly, fade, slide).
    *   **Hover**: Subtle brightness shifts or border glows, never just a rough color swap.

### 1.2 Theming
*   **Default**: Light Mode (Optimized for "Airy/Clean" feel).
*   **Dark Mode**: "Deep Space" feel (Deep blue/gray backgrounds, not pure black).
*   *Note: We prioritize Light Mode for the mass market "daytime travel" context, but Dark Mode must be fully supported.*

---

## 2. Technical Architecture (SvelteKit 5)

We use **Svelte 5** exclusively. Legacy options API or Svelte 4 patterns are strictly forbidden unless a library necessitates it.

### 2.1 Directory Structure
```
client/src/
├── lib/
│   ├── components/
│   │   ├── ui/           # Atomic Design: Base elements (shadcn-svelte + Custom Glass)
│   │   ├── blocks/       # Molecules: SearchBar, TicketCard, PaymentForm
│   │   └── layouts/      # Organisms: Navbar, Footer, Sidebar
│   ├── runes/            # Global state management using Svelte 5 runes (.svelte.ts)
│   ├── utils/            # Pure functions (formatters, validators)
│   └── styles/           # Global CSS and Tailwind variants
├── routes/
│   ├── (app)/            # Main authenticated app
│   ├── (marketing)/      # Landing, About, features
│   └── api/              # Internal proxy APIs
```

### 2.2 Component Rules
1.  **Single Responsibility**: A component does ONE thing. If it exceeds 200 lines, break it down.
2.  **Props Interface**: Always define props using `$props()`.
    ```typescript
    // ✅ svelte 5
    let { title, isActive = false, children } = $props<{
      title: string;
      isActive?: boolean;
      children?: Snippet;
    }>();
    ```
3.  **Snippets over Slots**: Use Svelte 5 `{#snippet}` instead of `<slot>`.

### 2.3 State Management
1.  **Local State**: Use `$state()` for component-local logic.
2.  **Shared State**: Use `.svelte.ts` files with `$state()` classes for global stores (e.g., `UserStore`, `CartStore`).
    *   *Do not* use Svelte 4 `writable()` stores unless integrating with legacy packages.
3.  **URL State**: For filter/search/pagination, **Always** sync with URL search params. The URL is the source of truth for navigation.

---

## 3. Coding Standards & Best Practices

### 3.1 TypeScript
*   **Strict Mode**: Enabled. No `any` types allowed. Define interfaces for all data models.
*   **API Types**: Generate types from the Go backend Protobuf specs or OpenAPI specs to ensure contract safety.

### 3.2 Performance Rules
1.  **Image Optimization**: Use `vite-imagetools` or Svelte optimized image components. All images above the fold must use `fetchpriority="high"`.
2.  **Lazy Loading**: Defer non-critical components using dynamic imports or `await` blocks.
3.  **CLS (Cumulative Layout Shift)**: Always reserve space for images/widgets skeleton states (Shadcn `Skeleton`).

### 3.3 Accessibility (a11y)
*   **Mandatory**: All interactive elements must have `aria-label` if no text is present.
*   **Keyboard**: All "clickable" `div`s must handle `onkeydown` and have `role="button"`.
*   **Contrast**: Ensure text passes WCAG AA against the glass backgrounds.

### 3.4 Library Specifics
*   **Icons (Lucide)**: Always import from `@lucide/svelte` (e.g., `import { Ticket } from "@lucide/svelte"`), NEVER `@lucide/svelte`.

---

## 4. UI Patterns & Snippets

### 4.1 The "Glass Card" Utility
Instead of raw CSS, we will standardize on a Tailwind utility class in `@layer components`:
```css
.glass-panel {
  @apply bg-white/70 dark:bg-slate-900/60 backdrop-blur-xl border border-white/20 dark:border-white/10 shadow-lg rounded-2xl;
}
```

### 4.2 Buttons
*   **Primary**: Gradient background (Brand Blue), Shadow-lg, `active:scale-95`.
*   **Secondary**: Glass style (Border, transparent bg), `hover:bg-white/20`.

### 4.3 Typography
*   **Font**: *Plus Jakarta Sans* (installed).
*   **Headings**: Bold, Tight tracking (`tracking-tight`).
*   **Body**: Readable, relaxed line-height.

---

## 5. Development Workflow

1.  **Atomic First**: build the small button/input/card first.
2.  **Mock Data**: Use `$state` with dummy JSON to prototype UI logic before connecting API.
3.  **Error Handling**: Every server load function must handle `try/catch` and return graceful error states to the UI.

---

**Signed off by**: *Antigravity Agent*
**For Project**: *Travio*
