/**
 * @type {import('tailwindcss').Config}
 */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'dnd-gold': '#d4af37',
        'dnd-dark': '#120f0d',
        'dnd-panel': '#1c1612',
        'dnd-accent': '#3a2f28',
        'dnd-paper': '#f4e4bc',
        'dnd-ink': '#2b1d0c',
        'dnd-sepia': '#5c4033',
      },
      fontFamily: {
        cinzel: ['Cinzel', 'serif'],
        medieval: ['MedievalSharp', 'cursive'],
        alegreya: ['Alegreya', 'serif'],
      },
    },
  },
  plugins: [],
}
