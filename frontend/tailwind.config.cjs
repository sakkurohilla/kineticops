/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#e6f2ff',
          100: '#cce5ff',
          200: '#99cbff',
          300: '#66b0ff',
          400: '#3396ff',
          500: '#007bff',
          600: '#0062cc',
          700: '#004a99',
          800: '#003166',
          900: '#001933',
        },
        secondary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        success: '#00d26a',
        warning: '#ff9800',
        error: '#f44336',
        dark: '#1a1a1a',
        light: '#f5f7fa',
        paytm: {
          blue: '#002970',
          lightBlue: '#00b9f5',
          darkBlue: '#001d4d',
        },
      },
      backgroundImage: {
        'gradient-paytm': 'linear-gradient(135deg, #002970 0%, #00b9f5 100%)',
        'gradient-card': 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
        'gradient-success': 'linear-gradient(135deg, #00d26a 0%, #00ff88 100%)',
        'gradient-warning': 'linear-gradient(135deg, #ff9800 0%, #ffb74d 100%)',
        'gradient-error': 'linear-gradient(135deg, #f44336 0%, #e57373 100%)',
      },
      boxShadow: {
        'paytm': '0 4px 20px rgba(0, 41, 112, 0.1)',
        'paytm-lg': '0 10px 40px rgba(0, 41, 112, 0.15)',
        'card': '0 2px 12px rgba(0, 0, 0, 0.08)',
      },
    },
  },
  plugins: [],
};
