<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { api } from '$lib/api';
	import {
		formatCost, formatBytes, formatDuration, formatNumber,
		formatTokens, formatDate, formatRelative, modelShortName, CHART_COLORS
	} from '$lib/utils';
	import type { OverviewResponse, SessionInfo, ProjectInfo, StorageInfo } from '$lib/types';

	Chart.register(...registerables);

	// ─── State ─────────────────────────────────────────────────────────────
	let activeTab = 'overview';
	let loading = true;
	let error = '';
	let lastRefresh = '';

	let overview: OverviewResponse | null = null;
	let sessions: SessionInfo[] = [];
	let projects: ProjectInfo[] = [];
	let storage: StorageInfo | null = null;
	let history: { totalEntries: number; commandCounts: Array<{command: string; count: number}> } | null = null;

	// Chart canvas refs
	let activityCanvas: HTMLCanvasElement;
	let tokenCanvas: HTMLCanvasElement;
	let modelCanvas: HTMLCanvasElement;
	let toolCanvas: HTMLCanvasElement;
	let storageCanvas: HTMLCanvasElement;

	let charts: Chart[] = [];
	let refreshInterval: ReturnType<typeof setInterval>;
	let sessionSearch = '';

	// ─── Data Loading ───────────────────────────────────────────────────────
	async function loadAll() {
		try {
			[overview, sessions, projects, storage, history] = await Promise.all([
				api.overview(),
				api.sessions(),
				api.projects(),
				api.storage(),
				api.history()
			]);
			lastRefresh = new Date().toLocaleTimeString();
			error = '';
		} catch (e) {
			error = String(e);
		} finally {
			loading = false;
		}
	}

	async function triggerRefresh() {
		await api.refresh();
		await new Promise(r => setTimeout(r, 2000));
		await loadAll();
		if (activeTab === 'overview') buildOverviewCharts();
		if (activeTab === 'tools') buildToolChart();
		if (activeTab === 'storage') buildStorageChart();
	}

	// ─── Charts ─────────────────────────────────────────────────────────────
	function destroyCharts() {
		charts.forEach(c => c.destroy());
		charts = [];
	}

	const chartDefaults = {
		responsive: true,
		maintainAspectRatio: false,
		plugins: {
			legend: { labels: { color: '#94a3b8', font: { size: 12 } } },
			tooltip: { backgroundColor: '#1a1d27', borderColor: '#2a2d3e', borderWidth: 1, titleColor: '#e2e8f0', bodyColor: '#94a3b8' }
		}
	};

	function buildOverviewCharts() {
		if (!overview) return;
		destroyCharts();

		const stats = overview.stats;

		// Daily Activity Line Chart
		if (activityCanvas) {
			const days = (stats.dailyActivity ?? []).slice(-60);
			const labels = days.map(d => {
				const dt = new Date(d.date);
				return `${dt.getMonth()+1}/${dt.getDate()}`;
			});
			charts.push(new Chart(activityCanvas, {
				type: 'line',
				data: {
					labels,
					datasets: [
						{
							label: 'Messages',
							data: days.map(d => d.messageCount),
							borderColor: '#7c3aed',
							backgroundColor: 'rgba(124,58,237,0.1)',
							fill: true,
							tension: 0.3,
							pointRadius: 2
						},
						{
							label: 'Tool Calls',
							data: days.map(d => d.toolCallCount),
							borderColor: '#06b6d4',
							backgroundColor: 'rgba(6,182,212,0.05)',
							fill: true,
							tension: 0.3,
							pointRadius: 2
						}
					]
				},
				options: {
					...chartDefaults,
					scales: {
						x: { ticks: { color: '#64748b', maxTicksLimit: 10 }, grid: { color: '#1e2130' } },
						y: { ticks: { color: '#64748b' }, grid: { color: '#1e2130' } }
					}
				}
			}));
		}

		// Token trend by model
		if (tokenCanvas) {
			const days = (stats.dailyModelTokens ?? []).slice(-60);
			const allModels = [...new Set(days.flatMap(d => Object.keys(d.tokensByModel ?? {})))];
			const labels = days.map(d => {
				const dt = new Date(d.date);
				return `${dt.getMonth()+1}/${dt.getDate()}`;
			});
			charts.push(new Chart(tokenCanvas, {
				type: 'bar',
				data: {
					labels,
					datasets: allModels.map((model, i) => ({
						label: modelShortName(model),
						data: days.map(d => Math.round((d.tokensByModel?.[model] ?? 0) / 1000)),
						backgroundColor: CHART_COLORS[i % CHART_COLORS.length] + 'cc',
						stack: 'tokens'
					}))
				},
				options: {
					...chartDefaults,
					scales: {
						x: { stacked: true, ticks: { color: '#64748b', maxTicksLimit: 10 }, grid: { color: '#1e2130' } },
						y: { stacked: true, ticks: { color: '#64748b', callback: (v) => `${v}K` }, grid: { color: '#1e2130' } }
					}
				}
			}));
		}

		// Model usage donut
		if (modelCanvas) {
			const usage = stats.modelUsage ?? {};
			const models = Object.keys(usage);
			const totals = models.map(m => {
				const u = usage[m];
				return u.inputTokens + u.outputTokens + u.cacheReadInputTokens + u.cacheCreationInputTokens;
			});
			charts.push(new Chart(modelCanvas, {
				type: 'doughnut',
				data: {
					labels: models.map(modelShortName),
					datasets: [{
						data: totals,
						backgroundColor: CHART_COLORS.slice(0, models.length),
						borderColor: '#1a1d27',
						borderWidth: 2
					}]
				},
				options: {
					...chartDefaults,
					cutout: '65%'
				}
			}));
		}
	}

	function buildToolChart() {
		if (!sessions.length || !toolCanvas) return;
		// Destroy only tool chart
		const idx = charts.findIndex(c => c.canvas === toolCanvas);
		if (idx !== -1) { charts[idx].destroy(); charts.splice(idx, 1); }

		const toolTotals: Record<string, number> = {};
		for (const s of sessions) {
			for (const [tool, count] of Object.entries(s.toolCalls ?? {})) {
				toolTotals[tool] = (toolTotals[tool] ?? 0) + count;
			}
		}
		const sorted = Object.entries(toolTotals).sort((a, b) => b[1] - a[1]).slice(0, 12);
		charts.push(new Chart(toolCanvas, {
			type: 'bar',
			data: {
				labels: sorted.map(([t]) => t),
				datasets: [{
					label: 'Total Calls',
					data: sorted.map(([, c]) => c),
					backgroundColor: CHART_COLORS.map(c => c + 'cc'),
					borderRadius: 4
				}]
			},
			options: {
				...chartDefaults,
				indexAxis: 'y' as const,
				plugins: { ...chartDefaults.plugins, legend: { display: false } },
				scales: {
					x: { ticks: { color: '#64748b' }, grid: { color: '#1e2130' } },
					y: { ticks: { color: '#e2e8f0', font: { size: 12 } }, grid: { display: false } }
				}
			}
		}));
	}

	function buildStorageChart() {
		if (!storage || !storageCanvas) return;
		const idx = charts.findIndex(c => c.canvas === storageCanvas);
		if (idx !== -1) { charts[idx].destroy(); charts.splice(idx, 1); }

		const dirs = Object.entries(storage.byDir).sort((a, b) => b[1] - a[1]).slice(0, 8);
		charts.push(new Chart(storageCanvas, {
			type: 'doughnut',
			data: {
				labels: dirs.map(([d]) => d),
				datasets: [{
					data: dirs.map(([, b]) => b),
					backgroundColor: CHART_COLORS.slice(0, dirs.length),
					borderColor: '#1a1d27',
					borderWidth: 2
				}]
			},
			options: { ...chartDefaults, cutout: '60%' }
		}));
	}

	// ─── Tab switching ───────────────────────────────────────────────────────
	function switchTab(tab: string) {
		activeTab = tab;
		// Build charts after DOM updates
		setTimeout(() => {
			if (tab === 'overview') buildOverviewCharts();
			if (tab === 'tools') buildToolChart();
			if (tab === 'storage') buildStorageChart();
		}, 50);
	}

	// ─── Lifecycle ───────────────────────────────────────────────────────────
	onMount(async () => {
		await loadAll();
		buildOverviewCharts();
		refreshInterval = setInterval(async () => {
			await loadAll();
			if (activeTab === 'overview') buildOverviewCharts();
			if (activeTab === 'tools') buildToolChart();
			if (activeTab === 'storage') buildStorageChart();
		}, 60_000);
	});

	onDestroy(() => {
		destroyCharts();
		clearInterval(refreshInterval);
	});

	// ─── Computed ────────────────────────────────────────────────────────────
	$: filteredSessions = sessions.filter(s =>
		!sessionSearch ||
		s.title?.toLowerCase().includes(sessionSearch.toLowerCase()) ||
		s.project?.toLowerCase().includes(sessionSearch.toLowerCase()) ||
		s.lastPrompt?.toLowerCase().includes(sessionSearch.toLowerCase())
	);

	$: totalToolCalls = sessions.reduce((sum, s) => sum + (s.totalToolCalls ?? 0), 0);
	$: totalErrors = sessions.reduce((s, x) => s + (x.errorCount ?? 0), 0);
	$: thinkingSessions = sessions.filter(s => s.hasThinking).length;
	$: toolTotals = (() => {
		const t: Record<string, number> = {};
		for (const s of sessions) for (const [tool, count] of Object.entries(s.toolCalls ?? {})) t[tool] = (t[tool] ?? 0) + count;
		return Object.entries(t).sort((a, b) => b[1] - a[1]);
	})();
	$: totalCacheHitRate = (() => {
		if (!overview?.stats?.modelUsage) return 0;
		let read = 0, input = 0;
		for (const u of Object.values(overview.stats.modelUsage)) {
			read += u.cacheReadInputTokens;
			input += u.inputTokens;
		}
		return read + input > 0 ? Math.round((read / (read + input)) * 100) : 0;
	})();

	$: hourData = (() => {
		const hc = overview?.stats?.hourCounts ?? {};
		return Array.from({ length: 24 }, (_, i) => hc[String(i)] ?? 0);
	})();
	$: maxHour = Math.max(...hourData, 1);

	const tabs = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'sessions', label: 'Sessions' },
		{ id: 'projects', label: 'Projects' },
		{ id: 'tools', label: 'Tools' },
		{ id: 'storage', label: 'Storage' },
		{ id: 'settings', label: 'Settings' }
	];
</script>

<div class="min-h-screen bg-surface text-slate-200">
	<!-- Header -->
	<header class="border-b border-surface-border px-6 py-4 flex items-center justify-between sticky top-0 z-10 bg-surface/95 backdrop-blur">
		<div class="flex items-center gap-3">
			<div class="w-7 h-7 rounded-lg bg-brand-purple flex items-center justify-center text-white font-bold text-sm">C</div>
			<h1 class="text-lg font-semibold tracking-tight">Claude Code Stats</h1>
		</div>
		<div class="flex items-center gap-4">
			{#if lastRefresh}
				<span class="text-xs text-slate-500">Updated {lastRefresh}</span>
			{/if}
			<button
				on:click={triggerRefresh}
				class="text-xs px-3 py-1.5 rounded-lg bg-surface-card border border-surface-border hover:border-brand-purple transition-colors"
			>
				↻ Refresh
			</button>
		</div>
	</header>

	<!-- Tab Navigation -->
	<nav class="border-b border-surface-border px-6 py-2 flex gap-1">
		{#each tabs as tab}
			<button
				class="tab-btn {activeTab === tab.id ? 'active' : ''}"
				on:click={() => switchTab(tab.id)}
			>
				{tab.label}
			</button>
		{/each}
	</nav>

	<!-- Content -->
	<main class="p-6 max-w-screen-2xl mx-auto">
		{#if loading}
			<div class="flex items-center justify-center h-64 text-slate-500">
				<div class="text-center">
					<div class="w-8 h-8 border-2 border-brand-purple border-t-transparent rounded-full animate-spin mx-auto mb-3"></div>
					<p>Parsing ~/.claude data…</p>
				</div>
			</div>
		{:else if error}
			<div class="card border-brand-red/40 text-brand-red">{error}</div>
		{:else}

		<!-- ═══════════════ OVERVIEW TAB ═══════════════ -->
		{#if activeTab === 'overview'}
			<!-- KPI Cards -->
			<div class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 mb-6">
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Sessions</p>
					<p class="text-2xl font-bold text-white">{formatNumber(overview?.stats?.totalSessions ?? 0)}</p>
					<p class="text-xs text-slate-500 mt-1">since {formatDate(overview?.stats?.firstSessionDate ?? '')}</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Messages</p>
					<p class="text-2xl font-bold text-white">{formatNumber(overview?.stats?.totalMessages ?? 0)}</p>
					<p class="text-xs text-slate-500 mt-1">{formatNumber(totalToolCalls)} tool calls</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Total Cost</p>
					<p class="text-2xl font-bold text-brand-green">{formatCost(overview?.totalCostUSD ?? 0)}</p>
					<p class="text-xs text-slate-500 mt-1">{projects.length} projects</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Cache Hit Rate</p>
					<p class="text-2xl font-bold text-brand-cyan">{totalCacheHitRate}%</p>
					<p class="text-xs text-slate-500 mt-1">tokens from cache</p>
				</div>
				<div class="card xl:col-span-2">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Thinking Speed</p>
					<p class="text-2xl font-bold text-brand-purple">{overview?.thinkingWPM ? Math.round(overview.thinkingWPM).toLocaleString() : '—'} <span class="text-sm font-normal text-slate-400">wpm</span></p>
					<p class="text-xs text-slate-500 mt-1">words Claude thinks per minute of API time</p>
				</div>
			</div>

			<!-- Charts Row 1 -->
			<div class="grid grid-cols-1 xl:grid-cols-3 gap-4 mb-4">
				<div class="card xl:col-span-2">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Daily Activity (last 60 days)</h2>
					<div class="h-56">
						<canvas bind:this={activityCanvas}></canvas>
					</div>
				</div>
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Model Usage</h2>
					<div class="h-56">
						<canvas bind:this={modelCanvas}></canvas>
					</div>
				</div>
			</div>

			<!-- Charts Row 2 -->
			<div class="grid grid-cols-1 xl:grid-cols-3 gap-4 mb-4">
				<div class="card xl:col-span-2">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Token Usage by Model (last 60 days, K tokens)</h2>
					<div class="h-56">
						<canvas bind:this={tokenCanvas}></canvas>
					</div>
				</div>
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Hour of Day (UTC)</h2>
					<div class="grid grid-cols-12 gap-0.5">
						{#each hourData as count, hour}
							<div class="flex flex-col items-center gap-0.5">
								<div
									class="w-full rounded-sm transition-all"
									style="height: {Math.max(4, (count / maxHour) * 60)}px; background: rgba(124,58,237,{count / maxHour})"
									title="{hour}:00 — {count} sessions"
								></div>
								{#if hour % 6 === 0}
									<span class="text-[8px] text-slate-600">{hour}h</span>
								{:else}
									<span class="text-[8px] text-transparent">·</span>
								{/if}
							</div>
						{/each}
					</div>
					<div class="mt-3 space-y-1">
						{#each Object.entries(overview?.computedCosts ?? {}) as [model, cost]}
							<div class="flex justify-between text-xs">
								<span class="text-slate-400">{modelShortName(model)}</span>
								<span class="text-slate-300 font-mono">{formatCost(cost)}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Longest session + model detail -->
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Longest Session</h2>
					{#if overview?.stats?.longestSession?.messageCount}
						<p class="text-2xl font-bold text-white mb-1">{formatNumber(overview.stats.longestSession.messageCount)} messages</p>
						<p class="text-sm text-slate-400">{formatDate(overview.stats.longestSession.timestamp)}</p>
					{:else}
						<p class="text-slate-500">No data</p>
					{/if}
				</div>
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Token Breakdown</h2>
					<div class="space-y-2">
						{#each Object.entries(overview?.stats?.modelUsage ?? {}) as [model, u]}
							<div>
								<div class="flex justify-between text-xs mb-1">
									<span class="text-slate-400">{modelShortName(model)}</span>
									<span class="text-slate-300">{formatTokens(u.inputTokens + u.outputTokens + u.cacheReadInputTokens)} tokens</span>
								</div>
								<div class="h-1.5 bg-surface-border rounded-full overflow-hidden flex gap-0.5">
									<div class="h-full bg-brand-purple rounded-full" style="width: {Math.round(u.inputTokens / Math.max(u.inputTokens + u.cacheReadInputTokens, 1) * 100)}%"></div>
									<div class="h-full bg-brand-cyan rounded-full" style="width: {Math.round(u.cacheReadInputTokens / Math.max(u.inputTokens + u.cacheReadInputTokens, 1) * 100)}%"></div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

		<!-- ═══════════════ SESSIONS TAB ═══════════════ -->
		{:else if activeTab === 'sessions'}
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold">{formatNumber(sessions.length)} Sessions</h2>
				<input
					bind:value={sessionSearch}
					placeholder="Search sessions…"
					class="px-3 py-1.5 rounded-lg bg-surface-card border border-surface-border text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-brand-purple w-64"
				/>
			</div>
			<div class="card overflow-hidden p-0">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-surface-border text-left">
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Title / Prompt</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Project</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Date</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Duration</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Msgs</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Tools</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Cost</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Flags</th>
						</tr>
					</thead>
					<tbody>
						{#each filteredSessions.slice(0, 100) as s}
							<tr class="border-b border-surface-border/50 hover:bg-surface-hover transition-colors">
								<td class="px-4 py-3 max-w-xs">
									{#if s.title}
										<p class="font-medium text-slate-200 truncate">{s.title}</p>
									{/if}
									{#if s.lastPrompt}
										<p class="text-xs text-slate-500 truncate mt-0.5">{s.lastPrompt}</p>
									{/if}
								</td>
								<td class="px-4 py-3 text-slate-400 whitespace-nowrap">{s.project || '—'}</td>
								<td class="px-4 py-3 text-slate-400 whitespace-nowrap text-xs">{formatRelative(s.startTime)}</td>
								<td class="px-4 py-3 text-slate-300 whitespace-nowrap">{formatDuration(s.durationMs)}</td>
								<td class="px-4 py-3 text-slate-300">{s.messageCount}</td>
								<td class="px-4 py-3 text-slate-300">{s.totalToolCalls}</td>
								<td class="px-4 py-3 font-mono text-brand-green whitespace-nowrap">{formatCost(s.costUSD)}</td>
								<td class="px-4 py-3">
									<div class="flex gap-1 flex-wrap">
										{#if s.hasThinking}
											<span class="badge bg-brand-purple/20 text-brand-purple">think</span>
										{/if}
										{#if s.errorCount > 0}
											<span class="badge bg-brand-red/20 text-brand-red">{s.errorCount}err</span>
										{/if}
										{#if s.fileCount > 0}
											<span class="badge bg-slate-700 text-slate-300">{s.fileCount}f</span>
										{/if}
										{#if s.version}
											<span class="badge bg-slate-700 text-slate-400">{s.version.replace('2.', '')}</span>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				{#if filteredSessions.length > 100}
					<p class="px-4 py-3 text-xs text-slate-500">Showing 100 of {filteredSessions.length} sessions</p>
				{/if}
			</div>

		<!-- ═══════════════ PROJECTS TAB ═══════════════ -->
		{:else if activeTab === 'projects'}
			<div class="mb-4">
				<h2 class="text-lg font-semibold">{projects.length} Projects</h2>
			</div>
			<div class="card overflow-hidden p-0 mb-4">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-surface-border text-left">
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Project</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Sessions</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Messages</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Tokens</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Cost</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Size</th>
							<th class="px-4 py-3 text-xs text-slate-500 font-medium">Last Active</th>
						</tr>
					</thead>
					<tbody>
						{#each projects as p}
							<tr class="border-b border-surface-border/50 hover:bg-surface-hover transition-colors">
								<td class="px-4 py-3">
									<p class="font-medium text-slate-200">{p.name}</p>
									<p class="text-xs text-slate-600 truncate max-w-xs">{p.path}</p>
								</td>
								<td class="px-4 py-3 text-slate-300">{p.sessionCount}</td>
								<td class="px-4 py-3 text-slate-300">{formatNumber(p.messageCount)}</td>
								<td class="px-4 py-3 text-slate-300">{formatTokens(p.totalTokens)}</td>
								<td class="px-4 py-3 font-mono text-brand-green">{formatCost(p.costUSD)}</td>
								<td class="px-4 py-3 text-slate-400">{formatBytes(p.sizeBytes)}</td>
								<td class="px-4 py-3 text-slate-400 text-xs">{formatRelative(p.lastActive)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
			<!-- Cost bar chart inline -->
			<div class="card">
				<h3 class="text-sm font-medium text-slate-400 mb-3">Cost by Project</h3>
				<div class="space-y-2">
					{#each projects.slice(0, 15) as p}
						{@const maxCost = Math.max(...projects.map(x => x.costUSD), 0.001)}
						<div class="flex items-center gap-3">
							<span class="text-xs text-slate-400 w-32 truncate text-right">{p.name}</span>
							<div class="flex-1 h-4 bg-surface-border rounded-full overflow-hidden">
								<div
									class="h-full bg-brand-purple rounded-full transition-all"
									style="width: {(p.costUSD / maxCost) * 100}%"
								></div>
							</div>
							<span class="text-xs font-mono text-brand-green w-16 text-right">{formatCost(p.costUSD)}</span>
						</div>
					{/each}
				</div>
			</div>

		<!-- ═══════════════ TOOLS TAB ═══════════════ -->
		{:else if activeTab === 'tools'}
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Total Tool Calls</p>
					<p class="text-3xl font-bold text-white">{formatNumber(totalToolCalls)}</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Unique Tools</p>
					<p class="text-3xl font-bold text-white">{toolTotals.length}</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Tool Errors</p>
					<p class="text-3xl font-bold text-brand-red">{formatNumber(totalErrors)}</p>
					<p class="text-xs text-slate-500 mt-1">{sessions.length > 0 ? ((totalErrors / sessions.length).toFixed(1)) : 0} per session</p>
				</div>
				<div class="card">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Extended Thinking</p>
					<p class="text-3xl font-bold text-brand-purple">{thinkingSessions}</p>
					<p class="text-xs text-slate-500 mt-1">{sessions.length > 0 ? Math.round((thinkingSessions / sessions.length) * 100) : 0}% of sessions</p>
				</div>
			</div>

			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Tool Call Frequency</h2>
					<div class="h-80">
						<canvas bind:this={toolCanvas}></canvas>
					</div>
				</div>
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Tool Breakdown</h2>
					<div class="space-y-2 overflow-y-auto max-h-80">
						{#each toolTotals as [tool, count]}
							{@const max = toolTotals[0]?.[1] ?? 1}
							<div class="flex items-center gap-3">
								<span class="text-xs text-slate-300 w-24 truncate font-mono">{tool}</span>
								<div class="flex-1 h-3 bg-surface-border rounded-full overflow-hidden">
									<div class="h-full bg-brand-cyan rounded-full" style="width: {(count / max) * 100}%"></div>
								</div>
								<span class="text-xs text-slate-400 w-12 text-right">{formatNumber(count)}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Slash command history -->
			{#if history?.commandCounts?.length}
				<div class="card mt-4">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Most Used Slash Commands</h2>
					<div class="flex flex-wrap gap-2">
						{#each history.commandCounts as { command, count }}
							<div class="card-sm flex items-center gap-2">
								<span class="text-brand-purple font-mono text-sm">{command}</span>
								<span class="badge bg-surface-border text-slate-400">{count}</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

		<!-- ═══════════════ STORAGE TAB ═══════════════ -->
		{:else if activeTab === 'storage'}
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
				<div class="card lg:col-span-1">
					<p class="text-xs text-slate-500 uppercase tracking-wide mb-1">Total ~/.claude Size</p>
					<p class="text-3xl font-bold text-white">{formatBytes(storage?.totalBytes ?? 0)}</p>
					<p class="text-xs text-slate-500 mt-1">{sessions.length} session files</p>
				</div>
				<div class="card lg:col-span-2">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Storage Breakdown</h2>
					<div class="h-44">
						<canvas bind:this={storageCanvas}></canvas>
					</div>
				</div>
			</div>

			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">By Directory</h2>
					<div class="space-y-2">
						{#each Object.entries(storage?.byDir ?? {}).sort((a, b) => b[1] - a[1]) as [dir, bytes]}
							{@const maxBytes = Math.max(...Object.values(storage?.byDir ?? {}), 1)}
							<div class="flex items-center gap-3">
								<span class="text-xs text-slate-400 w-36 truncate font-mono">{dir}</span>
								<div class="flex-1 h-3 bg-surface-border rounded-full overflow-hidden">
									<div class="h-full bg-brand-amber rounded-full" style="width: {(bytes / maxBytes) * 100}%"></div>
								</div>
								<span class="text-xs text-slate-300 w-16 text-right">{formatBytes(bytes)}</span>
							</div>
						{/each}
					</div>
				</div>
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Top Projects by Size</h2>
					<div class="space-y-2">
						{#each (storage?.byProject ?? []).slice(0, 15) as p}
							{@const maxBytes = Math.max(...(storage?.byProject ?? []).map(x => x.bytes), 1)}
							<div class="flex items-center gap-3">
								<span class="text-xs text-slate-400 w-28 truncate">{p.project}</span>
								<div class="flex-1 h-3 bg-surface-border rounded-full overflow-hidden">
									<div class="h-full bg-brand-blue rounded-full" style="width: {(p.bytes / maxBytes) * 100}%"></div>
								</div>
								<span class="text-xs text-slate-300 w-12 text-right">{formatBytes(p.bytes)}</span>
								<span class="text-xs text-slate-500 w-8 text-right">{p.sessions}s</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

		<!-- ═══════════════ SETTINGS TAB ═══════════════ -->
		{:else if activeTab === 'settings'}
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Backend Status</h2>
					<div class="space-y-2 text-sm">
						<div class="flex justify-between">
							<span class="text-slate-400">Sessions parsed</span>
							<span class="text-slate-200">{sessions.length}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-slate-400">Projects found</span>
							<span class="text-slate-200">{projects.length}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-slate-400">Cache hit rate</span>
							<span class="text-brand-cyan">{totalCacheHitRate}%</span>
						</div>
						<div class="flex justify-between">
							<span class="text-slate-400">Data last updated</span>
							<span class="text-slate-200">{lastRefresh}</span>
						</div>
					</div>
					<div class="mt-4">
						<button
							on:click={triggerRefresh}
							class="w-full px-4 py-2 rounded-lg bg-brand-purple text-white text-sm font-medium hover:bg-brand-purple/90 transition-colors"
						>
							Force Refresh Data
						</button>
					</div>
				</div>

				<div class="card">
					<h2 class="text-sm font-medium text-slate-400 mb-3">Model Usage Summary</h2>
					<div class="space-y-3">
						{#each Object.entries(overview?.stats?.modelUsage ?? {}) as [model, u]}
							<div class="card-sm">
								<div class="flex justify-between items-start mb-1">
									<span class="text-sm font-medium text-slate-200">{modelShortName(model)}</span>
									<span class="text-sm font-mono text-brand-green">{formatCost(overview?.computedCosts?.[model] ?? 0)}</span>
								</div>
								<div class="grid grid-cols-2 gap-x-4 text-xs text-slate-500">
									<span>Input: {formatTokens(u.inputTokens)}</span>
									<span>Output: {formatTokens(u.outputTokens)}</span>
									<span>Cache read: {formatTokens(u.cacheReadInputTokens)}</span>
									<span>Cache write: {formatTokens(u.cacheCreationInputTokens)}</span>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{/if}

		{/if}
	</main>
</div>
