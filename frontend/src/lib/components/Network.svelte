<script lang="ts">
	import { formatNumber, formatRelative } from '$lib/utils';
	import type { SessionInfo } from '$lib/types';

	export let sessions: SessionInfo[] = [];

	$: totalWebSearches  = sessions.reduce((s, x) => s + (x.webSearchCount ?? 0), 0);
	$: totalWebFetches   = sessions.reduce((s, x) => s + (x.webFetchCount ?? 0), 0);

	// Domain frequency
	$: topDomains = (() => {
		const counts: Record<string, number> = {};
		for (const s of sessions) {
			for (const d of s.webFetchDomains ?? []) {
				counts[d] = (counts[d] ?? 0) + 1;
			}
		}
		return Object.entries(counts).sort(([,a],[,b]) => b - a).slice(0, 15);
	})();

	// Search query frequency
	$: topQueries = (() => {
		const counts: Record<string, number> = {};
		for (const s of sessions) {
			for (const q of s.webSearchQueries ?? []) {
				counts[q] = (counts[q] ?? 0) + 1;
			}
		}
		return Object.entries(counts).sort(([,a],[,b]) => b - a).slice(0, 15);
	})();

	// Bash network command totals
	$: bashTotals = (() => {
		const counts: Record<string, number> = {};
		for (const s of sessions) {
			for (const [cmd, n] of Object.entries(s.bashNetworkCounts ?? {})) {
				counts[cmd] = (counts[cmd] ?? 0) + n;
			}
		}
		return Object.entries(counts).sort(([,a],[,b]) => b - a);
	})();

	$: bashTotal = bashTotals.reduce((s, [,n]) => s + n, 0);

	// Sessions with most network activity
	$: heavySessions = [...sessions]
		.map(s => ({
			...s,
			networkTotal: (s.webSearchCount ?? 0) + (s.webFetchCount ?? 0) +
				Object.values(s.bashNetworkCounts ?? {}).reduce((a,b) => a+b, 0)
		}))
		.filter(s => s.networkTotal > 0)
		.sort((a, b) => b.networkTotal - a.networkTotal)
		.slice(0, 10);

	$: maxDomain = topDomains[0]?.[1] ?? 1;
	$: maxBash   = bashTotals[0]?.[1] ?? 1;
</script>

<!-- KPI row -->
<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Web Searches</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(totalWebSearches)}</p>
		<p class="text-xs text-ink-faint mt-1">via WebSearch tool</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Web Fetches</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(totalWebFetches)}</p>
		<p class="text-xs text-ink-faint mt-1">via WebFetch tool</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Unique Domains</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(topDomains.length)}</p>
		<p class="text-xs text-ink-faint mt-1">fetched in total</p>
	</div>
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-2">Bash Network</p>
		<p class="text-2xl font-medium text-ink">{formatNumber(bashTotal)}</p>
		<p class="text-xs text-ink-faint mt-1">curl / installs / clones</p>
	</div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">

	<!-- Top domains -->
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Top Domains Fetched</p>
		{#if topDomains.length === 0}
			<p class="text-sm text-ink-muted">No web fetches recorded.</p>
		{:else}
			<div class="space-y-2.5">
				{#each topDomains as [domain, count]}
					<div>
						<div class="flex justify-between text-xs mb-1">
							<span class="text-ink-secondary truncate mr-2">{domain}</span>
							<span class="text-ink-muted shrink-0">{count}</span>
						</div>
						<div class="bar-track">
							<div class="bar-fill" style="width: {(count / maxDomain) * 100}%"></div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Bash network commands -->
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Bash Network Commands</p>
		{#if bashTotals.length === 0}
			<p class="text-sm text-ink-muted">No network bash commands recorded.</p>
		{:else}
			<div class="space-y-2.5">
				{#each bashTotals as [cmd, count]}
					<div>
						<div class="flex justify-between text-xs mb-1">
							<span class="font-mono text-ink-secondary">{cmd}</span>
							<span class="text-ink-muted">{count}</span>
						</div>
						<div class="bar-track">
							<div class="bar-fill" style="width: {(count / maxBash) * 100}%"></div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">

	<!-- Search queries -->
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Web Search Queries</p>
		{#if topQueries.length === 0}
			<p class="text-sm text-ink-muted">No web searches recorded.</p>
		{:else}
			<div class="space-y-1.5">
				{#each topQueries as [query, count]}
					<div class="flex items-start justify-between gap-2 py-1 border-b border-paper-border last:border-0">
						<span class="text-xs text-ink-secondary leading-relaxed">{query}</span>
						{#if count > 1}
							<span class="text-[10px] text-ink-faint shrink-0 mt-0.5">×{count}</span>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Sessions by network activity -->
	<div class="card">
		<p class="text-xs text-ink-muted tracking-wide mb-4">Most Active Sessions</p>
		{#if heavySessions.length === 0}
			<p class="text-sm text-ink-muted">No network activity found.</p>
		{:else}
			<div class="space-y-2">
				{#each heavySessions as s}
					<div class="flex items-center justify-between gap-2 py-1.5 border-b border-paper-border last:border-0">
						<div class="min-w-0">
							<p class="text-xs text-ink truncate">{s.title || s.project || '—'}</p>
							<p class="text-[10px] text-ink-faint mt-0.5">{formatRelative(s.startTime)}</p>
						</div>
						<div class="shrink-0 text-right">
							<p class="text-xs font-medium text-ink">{s.networkTotal} requests</p>
							<p class="text-[10px] text-ink-faint">
								{#if s.webFetchCount > 0}{s.webFetchCount} fetch{/if}
								{#if s.webFetchCount > 0 && s.webSearchCount > 0} · {/if}
								{#if s.webSearchCount > 0}{s.webSearchCount} search{/if}
							</p>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
