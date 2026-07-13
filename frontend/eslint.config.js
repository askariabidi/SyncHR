import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{js,jsx}'],
    extends: [
      js.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    rules: {
      // This one flags the standard "fetch data on mount" effect pattern
      // used throughout the app (useEffect(() => { fetchX() }, [])) as an
      // error. That's still the normal way to load data on mount in React
      // today, so it's downgraded to a warning rather than rewritten into
      // something more convoluted just to satisfy a very new, very strict
      // rule that most of the ecosystem hasn't caught up to yet.
      'react-hooks/set-state-in-effect': 'warn',
    },
  },
])
