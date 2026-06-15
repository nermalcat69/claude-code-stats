import type { OverviewResponse, SessionInfo, ProjectInfo, StorageInfo } from './types';

const BASE = '/api';

async function get<T>(path: string): Promise<T> {
	const res = await fetch(`${BASE}${path}`);
	if (!res.ok) throw new Error(`API error ${res.status}: ${path}`);
	return res.json();
}

export const api = {
	overview: () => get<OverviewResponse>('/overview'),
	sessions: () => get<SessionInfo[]>('/sessions'),
	projects: () => get<ProjectInfo[]>('/projects'),
	storage: () => get<StorageInfo>('/storage'),
	history: () => get<{ totalEntries: number; commandCounts: Array<{command: string; count: number}> }>('/history'),
	settings: () => get<{ settings: unknown; mcp: Record<string, unknown>; activeSessions: unknown[] }>('/settings'),
	status: () => get<{ status: string; ready: boolean; lastRefresh: string }>('/status'),
	refresh: () => fetch(`${BASE}/refresh`, { method: 'GET' })
};
