export interface DailyActivity {
	date: string;
	messageCount: number;
	userMessageCount: number;
	assistantMessageCount: number;
	sessionCount: number;
	toolCallCount: number;
}

export interface DailyModelTokens {
	date: string;
	tokensByModel: Record<string, number>;
}

export interface ModelUsageEntry {
	inputTokens: number;
	outputTokens: number;
	cacheReadInputTokens: number;
	cacheCreationInputTokens: number;
	webSearchRequests: number;
	costUSD: number;
}

export interface LongestSession {
	sessionId: string;
	duration: number;
	messageCount: number;
	userMessageCount: number;
	assistantMessageCount: number;
	linesAdded: number;
	timestamp: string;
}

export interface StatsCache {
	version: number;
	lastComputedDate: string;
	dailyActivity: DailyActivity[];
	dailyModelTokens: DailyModelTokens[];
	modelUsage: Record<string, ModelUsageEntry>;
	totalSessions: number;
	totalMessages: number;
	totalUserMessages: number;
	totalAssistantMessages: number;
	totalFilesEdited: number;
	totalLinesAdded: number;
	longestSession: LongestSession;
	firstSessionDate: string;
	hourCounts: Record<string, number>;
}

export interface TokenUsage {
	input_tokens: number;
	output_tokens: number;
	cache_read_input_tokens: number;
	cache_creation_input_tokens: number;
}

export interface SessionInfo {
	sessionId: string;
	project: string;
	projectPath: string;
	title: string;
	lastPrompt: string;
	startTime: string;
	endTime: string;
	durationMs: number;
	messageCount: number;
	userMessageCount: number;
	assistantMessageCount: number;
	toolCalls: Record<string, number>;
	totalToolCalls: number;
	costUSD: number;
	tokenUsage: TokenUsage;
	model: string;
	gitBranch: string;
	entryPoint: string;
	version: string;
	hasThinking: boolean;
	errorCount: number;
	fileCount: number;
	webSearchCount: number;
	webFetchCount: number;
	webSearchQueries: string[];
	webFetchDomains: string[];
	bashNetworkCounts: Record<string, number>;
}

export interface ProjectInfo {
	name: string;
	path: string;
	sessionCount: number;
	messageCount: number;
	userMessageCount: number;
	assistantMessageCount: number;
	totalTokens: number;
	costUSD: number;
	lastActive: string;
	sizeBytes: number;
}

export interface OverviewResponse {
	stats: StatsCache;
	computedCosts: Record<string, number>;
	totalCostUSD: number;
	thinkingWPM: number;
}

export interface StorageInfo {
	totalBytes: number;
	byDir: Record<string, number>;
	byProject: Array<{ project: string; bytes: number; sessions: number }>;
}
