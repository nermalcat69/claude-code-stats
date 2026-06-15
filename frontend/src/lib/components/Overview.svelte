<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { formatCost, formatNumber, formatTokens, formatDate, modelShortName, CHART_COLORS, chartBase } from '$lib/utils';
	import type { OverviewResponse } from '$lib/types';

	Chart.register(...registerables);

	export let overview: OverviewResponse | null = null;
	export let totalCacheHitRate = 0;

	let activityCanvas: HTMLCanvasElement;
	let tokenCanvas: HTMLCanvasElement;
	let modelCanvas: HTMLCanvasElement;
	let charts: Chart[] = [];
	let modelSortBy: 'tokens' | 'cost' = 'tokens';
	let chartPeriod = 60; // days; 0 = all time

	async function buildCharts() {
		await tick();
		charts.forEach(c => c.destroy());
		charts = [];
		if (!overview) return;

		const stats = overview.stats;

		if (activityCanvas) {
			const days = chartPeriod === 0
				? (stats.dailyActivity ?? [])
				: (stats.dailyActivity ?? []).slice(-chartPeriod);
			const fmtLabel = (iso: string) => {
				const [, m, d] = iso.split('-');
				return new Date(0, +m - 1, +d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
			};
			charts.push(new Chart(activityCanvas, {
				type: 'bar',
				data: {
					labels: days.map(d => fmtLabel(d.date)),
					datasets: [
						{
							label: 'Your messages',
							data: days.map(d => d.userMessageCount ?? 0),
							backgroundColor: '#D97757cc',
							borderRadius: 3,
							borderSkipped: false,
							stack: 'activity',
						},
						{
							label: 'Claude messages',
							data: days.map(d => d.assistantMessageCount ?? 0),
							backgroundColor: '#9B7FD4cc',
							borderRadius: 3,
							borderSkipped: false,
							stack: 'activity',
						},
						{
							label: 'Tool Calls',
							data: days.map(d => d.toolCallCount),
							backgroundColor: '#4A7C9Bcc',
							borderRadius: 3,
							borderSkipped: false,
							stack: 'activity',
						},
					],
				},
				options: {
					...chartBase,
					interaction: { mode: 'index', intersect: false },
					scales: {
						x: {
							stacked: true,
							...chartBase.scales.x,
							ticks: { ...chartBase.scales.x.ticks, maxTicksLimit: 12, maxRotation: 0 },
						},
						y: { stacked: true, ...chartBase.scales.y },
					},
					plugins: {
						...chartBase.plugins,
						tooltip: {
							...chartBase.plugins.tooltip,
							callbacks: {
								title: (items) => days[items[0].dataIndex]?.date ?? items[0].label,
								label: (item) => ` ${item.dataset.label}: ${item.formattedValue}`,
							},
						},
					},
				},
			}));
		}

		if (tokenCanvas) {
			const days = chartPeriod === 0
				? (stats.dailyModelTokens ?? [])
				: (stats.dailyModelTokens ?? []).slice(-chartPeriod);
			const models = [...new Set(days.flatMap(d => Object.keys(d.tokensByModel ?? {})))];
			const fmtLabel = (iso: string) => {
				const [, m, d] = iso.split('-');
				return new Date(0, +m - 1, +d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
			};
			charts.push(new Chart(tokenCanvas, {
				type: 'bar',
				data: {
					labels: days.map(d => fmtLabel(d.date)),
					datasets: models.map((m, i) => ({
						label: modelShortName(m),
						data: days.map(d => Math.round((d.tokensByModel?.[m] ?? 0) / 1000)),
						backgroundColor: CHART_COLORS[i % CHART_COLORS.length] + 'bb',
						stack: 'tokens',
						borderRadius: 2,
						borderSkipped: false,
					})),
				},
				options: {
					...chartBase,
					interaction: { mode: 'index', intersect: false },
					scales: {
						x: { stacked: true, ...chartBase.scales.x, ticks: { ...chartBase.scales.x.ticks, maxTicksLimit: 12, maxRotation: 0 } },
						y: { stacked: true, ...chartBase.scales.y, ticks: { ...chartBase.scales.y.ticks, callback: (v) => `${v}K` } },
					},
					plugins: {
						...chartBase.plugins,
						tooltip: {
							...chartBase.plugins.tooltip,
							callbacks: {
								title: (items) => days[items[0].dataIndex]?.date ?? items[0].label,
								label: (item) => ` ${item.dataset.label}: ${item.formattedValue}K tokens`,
							},
						},
					},
				},
			}));
		}

		if (modelCanvas) {
			const usage = stats.modelUsage ?? {};
			const mkeys = Object.keys(usage).sort((a, b) => {
				if (modelSortBy === 'cost') {
					const ca = overview?.computedCosts?.[a] ?? 0;
					const cb = overview?.computedCosts?.[b] ?? 0;
					return cb - ca;
				}
				const ta = usage[a].inputTokens + usage[a].outputTokens + usage[a].cacheReadInputTokens;
				const tb = usage[b].inputTokens + usage[b].outputTokens + usage[b].cacheReadInputTokens;
				return tb - ta;
			});
			const doughnutData = mkeys.map(m => {
				if (modelSortBy === 'cost') return overview?.computedCosts?.[m] ?? 0;
				const u = usage[m];
				return u.inputTokens + u.outputTokens + u.cacheReadInputTokens;
			});
			charts.push(new Chart(modelCanvas, {
				type: 'doughnut',
				data: {
					labels: mkeys.map(modelShortName),
					datasets: [{
						data: doughnutData,
						backgroundColor: CHART_COLORS.slice(0, mkeys.length),
						borderColor: '#FFFFFF',
						borderWidth: 3,
					}],
				},
				options: {
					...chartBase,
					cutout: '68%',
					scales: undefined as never,
					plugins: {
						...chartBase.plugins,
						tooltip: {
							...chartBase.plugins?.tooltip,
							callbacks: {
								label: (item) => {
									const val = modelSortBy === 'cost'
										? ` $${(item.raw as number).toFixed(4)}`
										: ` ${formatTokens(item.raw as number)} tokens`;
									return ` ${item.label}: ${val}`;
								},
							},
						},
					},
				},
			}));
		}
	}

	onMount(buildCharts);
	onDestroy(() => charts.forEach(c => c.destroy()));
	$: overview, modelSortBy, chartPeriod, buildCharts();

	$: daysActiveStats = (() => {
		const activity = overview?.stats?.dailyActivity ?? [];
		const cutoff = (n: number) => {
			const d = new Date();
			d.setDate(d.getDate() - n);
			return d.toISOString().split('T')[0];
		};
		return {
			total:   activity.length,
			last30:  activity.filter(d => d.date >= cutoff(30)).length,
			last120: activity.filter(d => d.date >= cutoff(120)).length,
		};
	})();

	$: chartPeriodLabel = chartPeriod === 0 ? 'all time' : `last ${chartPeriod} days`;

	$: hourData = (() => {
		const hc = overview?.stats?.hourCounts ?? {};
		return Array.from({ length: 24 }, (_, i) => hc[String(i)] ?? 0);
	})();
	$: maxHour = Math.max(...hourData, 1);

	$: sortedModelEntries = (() => {
		const usage = overview?.stats?.modelUsage ?? {};
		return Object.entries(usage).sort(([a], [b]) => {
			if (modelSortBy === 'cost') {
				return (overview?.computedCosts?.[b] ?? 0) - (overview?.computedCosts?.[a] ?? 0);
			}
			const ta = usage[a].inputTokens + usage[a].outputTokens + usage[a].cacheReadInputTokens;
			const tb = usage[b].inputTokens + usage[b].outputTokens + usage[b].cacheReadInputTokens;
			return tb - ta;
		});
	})();
</script>

<!-- Lifetime Stats -->
<p class="text-[10px] text-ink-muted tracking-widest uppercase mb-3">Lifetime Stats</p>
<div class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-3 mb-6">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Sessions</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(overview?.stats?.totalSessions ?? 0)}</p>
		<p class="text-xs text-ink-faint mt-1">since {formatDate(overview?.stats?.firstSessionDate ?? '')}</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Total Cost</p>
		<p class="text-2xl font-medium text-ink">{formatCost(overview?.totalCostUSD ?? 0)}</p>
		<p class="text-xs text-ink-faint mt-1">computed from tokens</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Cache Rate</p>
		<p class="text-2xl font-medium text-ink">{totalCacheHitRate}%</p>
		<p class="text-xs text-ink-faint mt-1">tokens from cache</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Thinking Speed</p>
		<p class="text-2xl font-medium text-terra">
			{overview?.thinkingWPM ? Math.round(overview.thinkingWPM).toLocaleString() : '—'}
			<span class="text-sm font-normal text-ink-secondary">wpm</span>
		</p>
		<p class="text-xs text-ink-faint mt-1">words / min of API time</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-3">Messages</p>
		<div class="space-y-2">
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Your messages</span>
				<span class="text-base font-medium text-ink">{formatNumber(overview?.stats?.totalUserMessages ?? 0)}</span>
			</div>
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Claude messages</span>
				<span class="text-base font-medium text-ink">{formatNumber(overview?.stats?.totalAssistantMessages ?? 0)}</span>
			</div>
		</div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-3">Code</p>
		<div class="space-y-2">
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Files edited</span>
				<span class="text-base font-medium text-ink">{formatNumber(overview?.stats?.totalFilesEdited ?? 0)}</span>
			</div>
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Lines added</span>
				<span class="text-base font-medium text-ink">{formatNumber(overview?.stats?.totalLinesAdded ?? 0)}</span>
			</div>
		</div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-3">Days Active</p>
		<div class="space-y-2">
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Lifetime</span>
				<span class="text-base font-medium text-ink">{daysActiveStats.total}</span>
			</div>
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Last 30d</span>
				<span class="text-base font-medium text-ink">{daysActiveStats.last30}<span class="text-xs text-ink-faint"> / 30</span></span>
			</div>
			<div class="flex items-baseline justify-between">
				<span class="text-xs text-ink-secondary">Last 120d</span>
				<span class="text-base font-medium text-ink">{daysActiveStats.last120}<span class="text-xs text-ink-faint"> / 120</span></span>
			</div>
		</div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-3">Longest Session</p>
		{#if overview?.stats?.longestSession?.messageCount}
			{@const ls = overview.stats.longestSession}
			<div class="space-y-2">
				<div class="flex items-baseline justify-between">
					<span class="text-xs text-ink-secondary">Your messages</span>
					<span class="text-base font-medium text-ink">{formatNumber(ls.userMessageCount)}</span>
				</div>
				<div class="flex items-baseline justify-between">
					<span class="text-xs text-ink-secondary">Claude messages</span>
					<span class="text-base font-medium text-ink">{formatNumber(ls.assistantMessageCount)}</span>
				</div>
				<div class="flex items-baseline justify-between border-t border-paper-border pt-2">
					<span class="text-xs text-ink-secondary">Lines added</span>
					<span class="text-base font-medium text-ink">{formatNumber(ls.linesAdded)}</span>
				</div>
				<p class="text-[10px] text-ink-faint">{formatDate(ls.timestamp)}</p>
			</div>
		{:else}
			<p class="text-ink-muted text-sm">No data</p>
		{/if}
	</div>
</div>

<!-- Period filter -->
<div class="flex items-center gap-2 mb-5">
	<span class="text-xs text-ink-muted mr-1">Chart period</span>
	{#each [{label: '30d', value: 30}, {label: '60d', value: 60}, {label: '90d', value: 90}, {label: '120d', value: 120}, {label: 'All', value: 0}] as p}
		<button
			class="chip {chartPeriod === p.value ? 'active' : ''}"
			on:click={() => chartPeriod = p.value}
		>{p.label}</button>
	{/each}
</div>

<!-- Charts row 1 -->
<div class="grid grid-cols-1 xl:grid-cols-3 gap-4 mb-4">
	<div class="card xl:col-span-2">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Daily Activity — {chartPeriodLabel}</p>
		<div class="h-52"><canvas bind:this={activityCanvas}></canvas></div>
	</div>
	<div class="card">
		<div class="flex items-center justify-between mb-4">
			<p class="text-xs text-ink-muted tracking-wide">Model Split</p>
			<div class="flex gap-1">
				<button
					class="text-[10px] px-2 py-0.5 rounded-full border transition-all duration-150 {modelSortBy === 'tokens' ? 'border-terra text-terra' : 'border-paper-border text-ink-faint hover:border-ink-secondary hover:text-ink-secondary'}"
					on:click={() => modelSortBy = 'tokens'}
				>tokens</button>
				<button
					class="text-[10px] px-2 py-0.5 rounded-full border transition-all duration-150 {modelSortBy === 'cost' ? 'border-terra text-terra' : 'border-paper-border text-ink-faint hover:border-ink-secondary hover:text-ink-secondary'}"
					on:click={() => modelSortBy = 'cost'}
				>cost</button>
			</div>
		</div>
		<div class="h-52"><canvas bind:this={modelCanvas}></canvas></div>
	</div>
</div>

<!-- Charts row 2 -->
<div class="grid grid-cols-1 xl:grid-cols-3 gap-4 mb-4">
	<div class="card xl:col-span-2">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Token Usage by Model — K tokens, {chartPeriodLabel}</p>
		<div class="h-52"><canvas bind:this={tokenCanvas}></canvas></div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Hour of Day (UTC)</p>
		<div class="flex items-end gap-0.5 h-16">
			{#each hourData as count, hour}
				<div
					class="flex-1 rounded-sm"
					style="height: {Math.max(4, (count / maxHour) * 100)}%; background: rgba(217,119,87,{0.15 + (count / maxHour) * 0.85})"
					title="{hour}:00 — {count} sessions"
				></div>
			{/each}
		</div>
		<div class="flex justify-between mt-1">
			{#each [0, 6, 12, 18, 23] as h}
				<span class="text-[9px] text-ink-faint">{h}h</span>
			{/each}
		</div>
		<div class="mt-4 space-y-1.5 border-t border-paper-border pt-4">
			{#each Object.entries(overview?.computedCosts ?? {}).sort(([,a],[,b]) => b - a) as [model, cost]}
				<div class="flex justify-between text-xs">
					<span class="text-ink-secondary">{modelShortName(model)}</span>
					<span class="text-ink font-medium">{formatCost(cost)}</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<!-- Bottom row -->
<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Cache Efficiency by Model</p>
		<div class="space-y-3">
			{#each sortedModelEntries as [model, u]}
				{@const total = u.inputTokens + u.cacheReadInputTokens}
				{@const cacheRate = total > 0 ? Math.round(u.cacheReadInputTokens / total * 100) : 0}
				<div>
					<div class="flex justify-between text-xs mb-1.5">
						<span class="text-ink-secondary">{modelShortName(model)}</span>
						<span class="text-ink-muted">{formatTokens(total)} total · {cacheRate}% cached</span>
					</div>
					<div class="bar-track">
						<div class="bar-fill" style="width: {cacheRate}%"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
