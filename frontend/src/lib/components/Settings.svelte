<script lang="ts">
	import { formatCost, formatTokens, modelShortName } from '$lib/utils';
	import type { OverviewResponse } from '$lib/types';

	export let overview: OverviewResponse | null = null;
	export let sessionCount = 0;
	export let projectCount = 0;
	export let totalCacheHitRate = 0;
	export let lastRefresh = '';
	export let onRefresh: () => void = () => {};
</script>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-5">Status</p>
		<div class="space-y-3">
			{#each [
				['Sessions parsed', sessionCount.toString()],
				['Projects found', projectCount.toString()],
				['Cache hit rate', `${totalCacheHitRate}%`],
				['Last updated', lastRefresh],
			] as [label, val]}
				<div class="flex items-center justify-between py-2 border-b border-paper-border last:border-0">
					<span class="text-sm text-ink-secondary">{label}</span>
					<span class="text-sm font-medium text-ink">{val}</span>
				</div>
			{/each}
		</div>
		<button
			on:click={onRefresh}
			class="mt-5 w-full px-4 py-2.5 rounded-full border border-paper-border bg-paper-card text-sm font-medium text-ink
			       hover:bg-ink hover:text-paper-card hover:border-ink transition-all duration-150"
		>
			Refresh data
		</button>
	</div>

	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-5">Model Usage</p>
		<div class="space-y-4">
			{#each Object.entries(overview?.stats?.modelUsage ?? {}) as [model, u]}
				<div class="pb-4 border-b border-paper-border last:border-0 last:pb-0">
					<div class="flex justify-between items-baseline mb-2">
						<span class="text-sm font-medium text-ink">{modelShortName(model)}</span>
						<span class="text-sm text-terra font-medium">{formatCost(overview?.computedCosts?.[model] ?? 0)}</span>
					</div>
					<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-ink-muted">
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
