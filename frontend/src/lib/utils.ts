export function formatCost(usd: number): string {
	if (usd >= 100) return `$${usd.toFixed(0)}`;
	if (usd >= 1)   return `$${usd.toFixed(2)}`;
	return `$${usd.toFixed(4)}`;
}

export function formatBytes(bytes: number): string {
	if (!bytes) return '0 B';
	const k = 1024, sizes = ['B', 'KB', 'MB', 'GB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

export function formatDuration(ms: number): string {
	if (ms <= 0) return '—';
	const m = Math.floor(ms / 60000), h = Math.floor(m / 60);
	if (h > 0) return `${h}h ${m % 60}m`;
	return m > 0 ? `${m}m` : `${Math.round(ms / 1000)}s`;
}

export function formatNumber(n: number): string { return n.toLocaleString(); }

export function formatTokens(n: number): string {
	if (n >= 1e9) return `${(n / 1e9).toFixed(1)}B`;
	if (n >= 1e6) return `${(n / 1e6).toFixed(1)}M`;
	if (n >= 1e3) return `${(n / 1e3).toFixed(1)}K`;
	return String(n);
}

export function formatDate(iso: string): string {
	if (!iso) return '—';
	return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function formatRelative(iso: string): string {
	if (!iso) return '—';
	const diff = Date.now() - new Date(iso).getTime();
	const d = Math.floor(diff / 86400000);
	if (d === 0) return 'Today';
	if (d === 1) return 'Yesterday';
	if (d < 7) return `${d}d ago`;
	if (d < 30) return `${Math.floor(d / 7)}w ago`;
	if (d < 365) return `${Math.floor(d / 30)}mo ago`;
	return `${Math.floor(d / 365)}y ago`;
}

export function modelShortName(model: string): string {
	const map: Record<string, string> = {
		'claude-sonnet-4-6':          'Sonnet 4.6',
		'claude-opus-4-6':            'Opus 4.6',
		'claude-opus-4-5-20251101':   'Opus 4.5',
		'claude-sonnet-4-5-20250929': 'Sonnet 4.5',
		'claude-haiku-4-5':           'Haiku 4.5',
		'claude-haiku-4-5-20251001':  'Haiku 4.5',
	};
	return map[model] ?? model.replace('claude-', '').replace(/-\d{8}$/, '');
}

/** Warm editorial palette for charts */
export const CHART_COLORS = [
	'#D97757', // terracotta
	'#4A7C9B', // dusty blue
	'#6B9E78', // sage green
	'#C4A463', // warm amber
	'#9B73A6', // muted lavender
	'#8A6B4A', // warm sienna
	'#5A8FA0', // steel blue
	'#7D8A5A', // olive
];

/** Shared Chart.js defaults for the warm aesthetic */
export const chartBase = {
	responsive: true,
	maintainAspectRatio: false,
	plugins: {
		legend: {
			labels: { color: '#77736D', font: { size: 11, family: 'Inter, sans-serif' } },
		},
		tooltip: {
			backgroundColor: '#FFFFFF',
			borderColor: '#E4E0D8',
			borderWidth: 1,
			titleColor: '#34322E',
			bodyColor: '#77736D',
			padding: 10,
			cornerRadius: 8,
		},
	},
	scales: {
		x: { ticks: { color: '#8C8882', font: { size: 10 } }, grid: { color: '#F0EDE8' } },
		y: { ticks: { color: '#8C8882', font: { size: 10 } }, grid: { color: '#F0EDE8' } },
	},
} as const;
