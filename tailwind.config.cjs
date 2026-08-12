/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50:  '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
          950: '#1e1b4b',
        },
        sidebar: {
          bg:          '#0f172a',
          hover:       '#1e293b',
          active:      '#1e293b',
          text:        '#94a3b8',
          active_text: '#f1f5f9',
          border:      '#1e293b',
        },
        surface: {
          DEFAULT: '#ffffff',
          50:  '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'monospace'],
      },
      boxShadow: {
        'card':    '0 1px 3px 0 rgba(0,0,0,.06), 0 1px 2px -1px rgba(0,0,0,.06)',
        'card-lg': '0 4px 6px -1px rgba(0,0,0,.07), 0 2px 4px -2px rgba(0,0,0,.07)',
        'glow':    '0 0 0 3px rgba(99,102,241,.25)',
        'login':   '0 25px 50px -12px rgba(0,0,0,.35)',
      },
      borderRadius: {
        '2.5xl': '1.25rem',
        '3xl':   '1.5rem',
        '4xl':   '2rem',
      },
      spacing: {
        'sidebar':           '260px',
        'sidebar-collapsed': '72px',
        'appbar':            '56px',
        'statusbar':         '24px',
      },
      zIndex: {
        'sidebar': '40',
        'appbar':  '30',
        'modal':   '50',
        'toast':   '9999',
        'palette': '60',
      },
      animation: {
        'fade-in':     'fadeIn .2s ease both',
        'slide-down':  'slideDown .25s ease both',
        'slide-up':    'slideUp .25s ease both',
        'slide-right': 'slideRight .25s ease both',
        'scale-in':    'scaleIn .15s ease both',
        'spin-slow':   'spin 2s linear infinite',
        'pulse-soft':  'pulse 2s cubic-bezier(.4,0,.6,1) infinite',
        'shimmer':     'shimmer 1.5s infinite',
      },
      keyframes: {
        fadeIn:     { from: { opacity: '0', transform: 'translateY(4px)' },  to: { opacity: '1', transform: 'translateY(0)' } },
        slideDown:  { from: { opacity: '0', transform: 'translateY(-8px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
        slideUp:    { from: { opacity: '0', transform: 'translateY(8px)' },  to: { opacity: '1', transform: 'translateY(0)' } },
        slideRight: { from: { opacity: '0', transform: 'translateX(-8px)' }, to: { opacity: '1', transform: 'translateX(0)' } },
        scaleIn:    { from: { opacity: '0', transform: 'scale(.95)' },       to: { opacity: '1', transform: 'scale(1)' } },
        shimmer:    { '0%': { backgroundPosition: '-200% 0' }, '100%': { backgroundPosition: '200% 0' } },
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
