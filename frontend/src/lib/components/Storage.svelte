<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { formatBytes, formatNumber, CHART_COLORS, chartBase } from '$lib/utils';
	import type { StorageInfo, SessionInfo } from '$lib/types';

	Chart.register(...registerables);

	export let storage: StorageInfo | null = null;
	export let sessions: SessionInfo[] = [];

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	$: maxDir     = Math.max(...Object.values(storage?.byDir ?? {}), 1);
	$: maxProject = Math.max(...(storage?.byProject ?? []).map(p => p.bytes), 1);

	async function buildChart() {
		await tick();
		chart?.destroy();
		if (!canvas || !storage) return;
		const dirs = Object.entries(storage.byDir).sort((a, b) => b[1] - a[1]).slice(0, 8);
		chart = new Chart(canvas, {
			type: 'doughnut',
			data: {
				labels: dirs.map(([d]) => d),
				datasets: [{
					data: dirs.map(([, b]) => b),
					backgroundColor: CHART_COLORS.slice(0, dirs.length),
					borderColor: '#FFFFFF',
					borderWidth: 3,
				}],
			},
			options: { ...chartBase, cutout: '62%', scales: undefined as never },
		});
	}

	onMount(buildChart);
	onDestroy(() => chart?.destroy());
	$: storage, buildChart();
</script>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Total Size</p>
		<p class="text-3xl font-medium text-ink">{formatBytes(storage?.totalBytes ?? 0)}</p>
		<p class="text-xs text-ink-faint mt-1">{formatNumber(sessions.length)} session files</p>
	</div>
	<div class="card lg:col-span-2">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Breakdown</p>
		<div class="h-40"><canvas bind:this={canvas}></canvas></div>
	</div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">By Directory</p>
		<div class="space-y-2.5">
			{#each Object.entries(storage?.byDir ?? {}).sort((a, b) => b[1] - a[1]) as [dir, bytes]}
				<div class="flex items-center gap-3">
					<span class="text-xs text-ink-secondary w-36 truncate font-mono">{dir}</span>
					<div class="bar-track">
						<div class="h-full rounded-full bg-terra transition-all" style="width: {(bytes / maxDir) * 100}%"></div>
					</div>
					<span class="text-xs text-ink w-14 text-right">{formatBytes(bytes)}</span>
				</div>
			{/each}
		</div>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Top Projects by Size</p>
		<div class="space-y-2.5">
			{#each (storage?.byProject ?? []).slice(0, 15) as p}
				<div class="flex items-center gap-3">
					<span class="text-xs text-ink-secondary w-28 truncate">{p.project}</span>
					<div class="bar-track">
						<div class="h-full rounded-full bg-[#4A7C9B] transition-all" style="width: {(p.bytes / maxProject) * 100}%"></div>
					</div>
					<span class="text-xs text-ink w-12 text-right">{formatBytes(p.bytes)}</span>
					<span class="text-xs text-ink-faint w-6 text-right">{p.sessions}s</span>
				</div>
			{/each}
		</div>
	</div>
</div>
