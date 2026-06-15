/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			fontFamily: {
				newsreader: ['Newsreader', 'Georgia', 'serif'],
				sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
			},
			colors: {
				paper: {
					DEFAULT: '#F7F7F5',
					card:    '#FFFFFF',
					border:  '#E4E0D8',
					hover:   '#F0EDE8',
					deep:    '#EAE6E0',
				},
				ink: {
					DEFAULT:   '#34322E',
					secondary: '#77736D',
					muted:     '#8C8882',
					faint:     '#B8B2AA',
				},
				terra: {
					DEFAULT: '#D97757',
					light:   '#F5E8E0',
					dark:    '#B85E3E',
				},
			},
		},
	},
	plugins: [],
};
