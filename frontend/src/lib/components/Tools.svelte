<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { formatNumber, CHART_COLORS, chartBase } from '$lib/utils';
	import type { SessionInfo } from '$lib/types';

	Chart.register(...registerables);

	export let sessions: SessionInfo[] = [];
	export let history: { totalEntries: number; commandCounts: Array<{command: string; count: number}> } | null = null;

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	$: totalToolCalls = sessions.reduce((s, x) => s + (x.totalToolCalls ?? 0), 0);
	$: totalErrors    = sessions.reduce((s, x) => s + (x.errorCount ?? 0), 0);
	$: thinkCount     = sessions.filter(s => s.hasThinking).length;

	$: toolTotals = (() => {
		const t: Record<string, number> = {};
		for (const s of sessions)
			for (const [tool, n] of Object.entries(s.toolCalls ?? {}))
				t[tool] = (t[tool] ?? 0) + n;
		return Object.entries(t).sort((a, b) => b[1] - a[1]);
	})();

	$: maxTool = toolTotals[0]?.[1] ?? 1;

	async function buildChart() {
		await tick();
		chart?.destroy();
		if (!canvas || !toolTotals.length) return;
		const top = toolTotals.slice(0, 12);
		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels: top.map(([t]) => t),
				datasets: [{
					label: 'Calls',
					data: top.map(([, n]) => n),
					backgroundColor: CHART_COLORS.map(c => c + 'cc'),
					borderRadius: 3,
					borderSkipped: false,
				}],
			},
			options: {
				...chartBase,
				indexAxis: 'y' as const,
				plugins: { ...chartBase.plugins, legend: { display: false } },
				scales: {
					x: { ...chartBase.scales.x },
					y: { ...chartBase.scales.y, grid: { display: false }, ticks: { color: '#34322E', font: { size: 11 } } },
				},
			},
		});
	}

	onMount(buildChart);
	onDestroy(() => chart?.destroy());
	$: sessions, buildChart();
</script>

<!-- KPI row -->
<div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Total Calls</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(totalToolCalls)}</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Tool Types</p>
		<p class="text-2xl font-medium text-ink">{toolTotals.length}</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Tool Errors</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(totalErrors)}</p>
		<p class="text-xs text-ink-faint mt-1">
			{sessions.length > 0 ? (totalErrors / sessions.length).toFixed(1) : 0} per session
		</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Extended Thinking</p>
		<p class="text-2xl font-medium text-terra">{thinkCount}</p>
		<p class="text-xs text-ink-faint mt-1">
			{sessions.length > 0 ? Math.round((thinkCount / sessions.length) * 100) : 0}% of sessions
		</p>
	</div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Call Frequency</p>
		<div class="h-72"><canvas bind:this={canvas}></canvas></div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">All Tools</p>
		<div class="space-y-2 overflow-y-auto max-h-72">
			{#each toolTotals as [tool, count]}
				<div class="flex items-center gap-3">
					<span class="text-xs text-ink-secondary w-24 truncate font-mono">{tool}</span>
					<div class="bar-track">
						<div class="h-full rounded-full bg-terra transition-all" style="width: {(count / maxTool) * 100}%"></div>
					</div>
					<span class="text-xs text-ink-muted w-12 text-right">{formatNumber(count)}</span>
				</div>
			{/each}
		</div>
	</div>
</div>

{#if history?.commandCounts?.length}
	<div class="card mt-4">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Slash Commands</p>
		<div class="flex flex-wrap gap-2">
			{#each history.commandCounts as { command, count }}
				<div class="flex items-center gap-2 px-3 py-1.5 rounded-full border border-paper-border bg-paper text-sm">
					<span class="text-terra font-medium">{command}</span>
					<span class="text-ink-muted">{count}</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
