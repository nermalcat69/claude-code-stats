<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import Overview  from '$lib/components/Overview.svelte';
	import Sessions  from '$lib/components/Sessions.svelte';
	import Projects  from '$lib/components/Projects.svelte';
	import Tools     from '$lib/components/Tools.svelte';
	import Network   from '$lib/components/Network.svelte';
	import Storage   from '$lib/components/Storage.svelte';
	import Settings  from '$lib/components/Settings.svelte';
	import type { OverviewResponse, SessionInfo, ProjectInfo, StorageInfo } from '$lib/types';

	// ─── State ────────────────────────────────────────────────────────────────
	let tab = 'overview';
	let loading = true;
	let error = '';
	let lastRefresh = '';

	let overview: OverviewResponse | null = null;
	let sessions: SessionInfo[] = [];
	let projects: ProjectInfo[] = [];
	let storage:  StorageInfo | null = null;
	let history:  { totalEntries: number; commandCounts: Array<{command: string; count: number}> } | null = null;

	let ticker: ReturnType<typeof setInterval>;

	// ─── Data ─────────────────────────────────────────────────────────────────
	async function load() {
		try {
			[overview, sessions, projects, storage, history] = await Promise.all([
				api.overview(), api.sessions(), api.projects(), api.storage(), api.history(),
			]);
			lastRefresh = new Date().toLocaleTimeString();
			error = '';
		} catch (e) {
			error = String(e);
		} finally {
			loading = false;
		}
	}

	async function refresh() {
		await api.refresh();
		await new Promise(r => setTimeout(r, 1800));
		await load();
	}

	onMount(() => { load(); ticker = setInterval(load, 60_000); });
	onDestroy(() => clearInterval(ticker));

	// ─── Derived ──────────────────────────────────────────────────────────────
	$: totalCacheHitRate = (() => {
		if (!overview?.stats?.modelUsage) return 0;
		let read = 0, input = 0;
		for (const u of Object.values(overview.stats.modelUsage)) {
			read  += u.cacheReadInputTokens;
			input += u.inputTokens;
		}
		return read + input > 0 ? Math.round((read / (read + input)) * 100) : 0;
	})();

	// ─── Greeting ─────────────────────────────────────────────────────────────
	const hour = new Date().getHours();
	const greeting = hour < 12 ? 'Good morning'
	               : hour < 17 ? 'Good afternoon'
	               : 'Good evening';

	const tabs = [
		{ id: 'overview',  label: 'Overview'  },
		{ id: 'sessions',  label: 'Sessions'  },
		{ id: 'projects',  label: 'Projects'  },
		{ id: 'tools',     label: 'Tools'     },
		{ id: 'network',   label: 'Network'   },
		{ id: 'storage',   label: 'Storage'   },
		{ id: 'settings',  label: 'Settings'  },
	];
</script>

<div class="min-h-screen bg-paper">

	<!-- ── Hero ─────────────────────────────────────────────────────────────── -->
	<div class="text-center pt-14 pb-10 px-6">
		<h1 class="font-newsreader text-[68px] leading-[1.1] font-normal text-ink">
			<span class="text-terra mr-2 align-baseline">✦</span>{greeting}.
		</h1>
		<p class="text-ink-secondary mt-3 text-base">Your Claude activity, at a glance.</p>

		<!-- Tab chips -->
		<nav class="flex justify-center gap-2 mt-8 flex-wrap">
			{#each tabs as t}
				<button
					class="chip {tab === t.id ? 'active' : ''}"
					on:click={() => (tab = t.id)}
				>{t.label}</button>
			{/each}
		</nav>
	</div>

	<!-- ── Content ───────────────────────────────────────────────────────────── -->
	<main class="px-6 pb-16 max-w-screen-xl mx-auto">

		{#if loading}
			<div class="flex flex-col items-center justify-center h-64 gap-4">
				<div class="w-6 h-6 rounded-full border-2 border-terra border-t-transparent animate-spin"></div>
				<p class="text-sm text-ink-muted">Parsing ~/.claude data…</p>
			</div>

		{:else if error}
			<div class="card border-red-200 text-red-700 text-sm">{error}</div>

		{:else if tab === 'overview'}
			<Overview {overview} {totalCacheHitRate} />

		{:else if tab === 'sessions'}
			<Sessions {sessions} />

		{:else if tab === 'projects'}
			<Projects {projects} />

		{:else if tab === 'tools'}
			<Tools {sessions} {history} />

		{:else if tab === 'network'}
			<Network {sessions} />

		{:else if tab === 'storage'}
			<Storage {storage} {sessions} />

		{:else if tab === 'settings'}
			<Settings
				{overview}
				sessionCount={sessions.length}
				projectCount={projects.length}
				{totalCacheHitRate}
				{lastRefresh}
				onRefresh={refresh}
			/>
		{/if}

	</main>
</div>
