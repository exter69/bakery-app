# No Emoji Policy

- NEVER use emoji characters (Unicode emoji) in code, components, templates, or UI strings.
- Use icon components (SVG icons, icon libraries like Lucide, Heroicons, or custom SVG) instead of emoji for visual indicators.
- This applies to:
  - React component JSX output
  - Translation strings (i18n)
  - CSS pseudo-elements
  - Comments that end up in user-facing output
  - Placeholder text in components
- If a design or mockup shows an emoji (e.g., 🌱, 🗺️, 🥐), replace it with an appropriate SVG icon or icon component.
- For decorative indicators (badges, status icons, category markers), use `<svg>` inline icons or an icon library.
