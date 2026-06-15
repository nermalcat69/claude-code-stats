<script lang="ts">
	import { formatCost, formatBytes, formatTokens, formatNumber, formatRelative } from '$lib/utils';
	import type { ProjectInfo } from '$lib/types';

	export let projects: ProjectInfo[] = [];

	$: maxCost = Math.max(...projects.map(p => p.costUSD), 0.0001);
</script>

<p class="text-sm text-ink-secondary mb-5">{projects.length} projects</p>

<div class="card p-0 overflow-hidden mb-4">
	<table class="w-full text-sm">
		<thead>
			<tr class="border-b border-paper-border">
				<th class="th">Project</th>
				<th class="th">Sessions</th>
				<th class="th">Your msgs</th>
				<th class="th">Claude msgs</th>
				<th class="th">Tokens</th>
				<th class="th">Cost</th>
				<th class="th">Size</th>
				<th class="th">Last active</th>
			</tr>
		</thead>
		<tbody>
			{#each projects as p}
				<tr class="trow">
					<td class="td">
						<p class="font-medium text-ink">{p.name}</p>
						<p class="text-xs text-ink-faint truncate max-w-xs">{p.path}</p>
					</td>
					<td class="td text-ink-secondary">{p.sessionCount}</td>
					<td class="td text-ink-secondary">{formatNumber(p.userMessageCount)}</td>
					<td class="td text-ink-secondary">{formatNumber(p.assistantMessageCount)}</td>
					<td class="td text-ink-secondary">{formatTokens(p.totalTokens)}</td>
					<td class="td font-medium text-ink">{formatCost(p.costUSD)}</td>
					<td class="td text-ink-muted">{formatBytes(p.sizeBytes)}</td>
					<td class="td text-ink-muted text-xs">{formatRelative(p.lastActive)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<!-- Cost bars -->
<div class="card">
	<p class="text-xs text-ink-muted tracking-wide mb-5">Cost by Project</p>
	<div class="space-y-3">
		{#each projects.slice(0, 15) as p}
			<div class="flex items-center gap-3">
				<span class="text-xs text-ink-secondary w-28 truncate text-right">{p.name}</span>
				<div class="bar-track">
					<div class="bar-fill" style="width: {(p.costUSD / maxCost) * 100}%"></div>
				</div>
				<span class="text-xs font-medium text-ink w-14 text-right">{formatCost(p.costUSD)}</span>
			</div>
		{/each}
	</div>
</div>
