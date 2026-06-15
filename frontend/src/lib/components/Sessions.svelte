<script lang="ts">
	import { formatCost, formatDate, formatDuration, formatNumber, formatRelative } from '$lib/utils';
	import type { SessionInfo } from '$lib/types';

	export let sessions: SessionInfo[] = [];

	let search = '';

	$: filtered = sessions.filter(s =>
		!search ||
		s.title?.toLowerCase().includes(search.toLowerCase()) ||
		s.project?.toLowerCase().includes(search.toLowerCase()) ||
		s.lastPrompt?.toLowerCase().includes(search.toLowerCase())
	);
</script>

<div class="flex items-center justify-between mb-5">
	<p class="text-sm text-ink-secondary">{formatNumber(sessions.length)} sessions</p>
	<input
		bind:value={search}
		placeholder="Search sessions…"
		class="w-64 px-4 py-2 rounded-full border border-paper-border bg-paper-card text-sm text-ink
		       placeholder-ink-faint focus:outline-none focus:border-ink-secondary transition-colors duration-150"
	/>
</div>

<div class="card p-0 overflow-hidden">
	<table class="w-full text-sm">
		<thead>
			<tr class="border-b border-paper-border">
				<th class="th">Session</th>
				<th class="th">Project</th>
				<th class="th">When</th>
				<th class="th">Duration</th>
				<th class="th">Msgs</th>
				<th class="th">Tools</th>
				<th class="th">Cost</th>
				<th class="th">Tags</th>
			</tr>
		</thead>
		<tbody>
			{#each filtered as s}
				<tr class="trow">
					<td class="td max-w-xs">
						<p class="font-medium text-ink truncate">{s.title || '—'}</p>
						{#if s.lastPrompt}
							<p class="text-xs text-ink-muted truncate mt-0.5">{s.lastPrompt}</p>
						{/if}
					</td>
					<td class="td whitespace-nowrap">
						<p class="text-ink-secondary">{s.project || '—'}</p>
						{#if s.projectPath}
							<p class="text-[10px] text-ink-faint mt-0.5 font-mono truncate max-w-[180px]">{s.projectPath}</p>
						{/if}
					</td>
					<td class="td whitespace-nowrap">
						<p class="text-xs text-ink-muted">{formatRelative(s.startTime)}</p>
						<p class="text-[10px] text-ink-faint mt-0.5">{formatDate(s.startTime)}</p>
					</td>
					<td class="td text-ink whitespace-nowrap">{formatDuration(s.durationMs)}</td>
					<td class="td text-ink-secondary">{s.messageCount}</td>
					<td class="td text-ink-secondary">{s.totalToolCalls}</td>
					<td class="td font-medium whitespace-nowrap text-ink">{formatCost(s.costUSD)}</td>
					<td class="td">
						<div class="flex gap-1 flex-wrap">
							{#if s.hasThinking}
								<span class="badge bg-terra-light text-terra">think</span>
							{/if}
							{#if s.errorCount > 0}
								<span class="badge bg-red-50 text-red-600">{s.errorCount}err</span>
							{/if}
							{#if s.fileCount > 0}
								<span class="badge bg-paper-deep text-ink-secondary">{s.fileCount}f</span>
							{/if}
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
