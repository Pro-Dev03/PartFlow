# PartFlow Design System

This folder is the single home for shared visual design code.

- `components/`: shared UI primitives and reusable visual components.
- `styles/`: design tokens, themes, global styles, and responsive rules.
- `index.ts`: public component entry point for new imports.

`src/index.css` remains the Vite stylesheet entry point and imports the files from `styles/`. `tailwind.config.js` remains at the frontend root because Tailwind resolves it from the project root. Fonts that must be served as public URLs remain in `public/fonts/`.
