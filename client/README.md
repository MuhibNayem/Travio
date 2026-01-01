# Travio Client

A modern SvelteKit 5 application with shadcn-svelte and Tailwind CSS 4.

## Tech Stack

- **[SvelteKit 5](https://kit.svelte.dev/)** - The latest version with Svelte 5 runes
- **[shadcn-svelte](https://shadcn-svelte.com/)** - Beautiful, accessible component library
- **[Tailwind CSS 4](https://tailwindcss.com/)** - Latest Tailwind with improved performance
- **[TypeScript](https://www.typescriptlang.org/)** - Type-safe development
- **[pnpm](https://pnpm.io/)** - Fast, disk space efficient package manager

## Svelte 5 Features

This project uses the latest Svelte 5 runes:

### `$state` - Reactive State
```svelte
let count = $state(0);
```

### `$derived` - Computed Values
```svelte
let doubleCount = $derived(count * 2);
let isEven = $derived(count % 2 === 0);
```

### `$effect` - Side Effects
```svelte
$effect(() => {
  console.log(`Count changed to: ${count}`);
});
```

### `$props` - Component Props
```svelte
let { children } = $props();
```

## Getting Started

### Prerequisites

- Node.js 18+ 
- pnpm (or npm/yarn)

### Installation

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Project Structure

```
client/
├── src/
│   ├── lib/
│   │   ├── components/    # Reusable components
│   │   │   └── ui/       # shadcn-svelte components
│   │   ├── utils/        # Utility functions
│   │   └── hooks/        # Custom hooks
│   └── routes/           # SvelteKit routes
│       ├── +layout.svelte
│       ├── +page.svelte
│       └── layout.css    # Global styles with Tailwind
├── static/               # Static assets
├── components.json       # shadcn-svelte configuration
├── svelte.config.js      # SvelteKit configuration
└── tsconfig.json         # TypeScript configuration
```

## Adding shadcn-svelte Components

You can add any component from the shadcn-svelte library:

```bash
npx shadcn-svelte@latest add button
npx shadcn-svelte@latest add card
npx shadcn-svelte@latest add dialog
# ... and many more
```

See all available components at [shadcn-svelte.com](https://shadcn-svelte.com/)

## Configuration

### Tailwind CSS

The project uses Tailwind CSS 4 with a custom theme configured in `src/routes/layout.css`. The theme includes:

- Custom color palette (slate base color)
- Dark mode support
- CSS variables for theming
- Custom radius values

### TypeScript

TypeScript is configured for strict type checking. See `tsconfig.json` for the configuration.

### Import Aliases

- `$lib` - src/lib
- `$lib/components` - src/lib/components
- `$lib/components/ui` - src/lib/components/ui (shadcn-svelte components)
- `$lib/utils` - src/lib/utils
- `$lib/hooks` - src/lib/hooks

## Development

### Code Quality

```bash
# Type checking
pnpm check

# Type checking with watch mode
pnpm check:watch
```

### Building

```bash
# Build for production
pnpm build

# Preview the production build
pnpm preview
```

## Learn More

- [SvelteKit Documentation](https://kit.svelte.dev/docs)
- [Svelte 5 Runes Documentation](https://svelte.dev/docs/svelte/$state)
- [shadcn-svelte Documentation](https://shadcn-svelte.com/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)

## License

Private - All rights reserved
