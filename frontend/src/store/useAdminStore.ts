import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { adminService } from '@/services/adminService';
import type {
  OverviewStats,
  ProviderAnalytics,
  ModelAnalytics,
  ChatAnalytics,
  AgentAnalytics,
} from '@/types/admin';

interface AdminState {
  // Analytics data
  overviewStats: OverviewStats | null;
  providerAnalytics: ProviderAnalytics[];
  modelAnalytics: ModelAnalytics[];
  chatAnalytics: ChatAnalytics | null;
  agentAnalytics: AgentAnalytics | null;

  // Loading states
  isLoadingStats: boolean;
  isLoadingProviders: boolean;
  isLoadingModels: boolean;
  isLoadingChats: boolean;
  isLoadingAgents: boolean;

  // Errors
  statsError: string | null;
  providersError: string | null;
  modelsError: string | null;
  chatsError: string | null;
  agentsError: string | null;

  // Actions
  fetchOverviewStats: () => Promise<void>;
  fetchProviderAnalytics: () => Promise<void>;
  fetchModelAnalytics: () => Promise<void>;
  fetchChatAnalytics: () => Promise<void>;
  fetchAgentAnalytics: () => Promise<void>;
  refreshAllAnalytics: () => Promise<void>;
}

export const useAdminStore = create<AdminState>()(
  devtools(
    set => ({
      // Initial state
      overviewStats: null,
      providerAnalytics: [],
      modelAnalytics: [],
      chatAnalytics: null,
      agentAnalytics: null,

      isLoadingStats: false,
      isLoadingProviders: false,
      isLoadingModels: false,
      isLoadingChats: false,
      isLoadingAgents: false,

      statsError: null,
      providersError: null,
      modelsError: null,
      chatsError: null,
      agentsError: null,

      // Actions
      fetchOverviewStats: async () => {
        set({ isLoadingStats: true, statsError: null });
        try {
          const stats = await adminService.getOverviewStats();
          set({ overviewStats: stats, isLoadingStats: false });
        } catch (error) {
          console.error('Failed to fetch overview stats:', error);
          set({
            statsError: error instanceof Error ? error.message : 'Failed to fetch stats',
            isLoadingStats: false,
          });
        }
      },

      fetchProviderAnalytics: async () => {
        set({ isLoadingProviders: true, providersError: null });
        try {
          const analytics = await adminService.getProviderAnalytics();
          set({ providerAnalytics: analytics, isLoadingProviders: false });
        } catch (error) {
          console.error('Failed to fetch provider analytics:', error);
          set({
            providersError:
              error instanceof Error ? error.message : 'Failed to fetch provider analytics',
            isLoadingProviders: false,
          });
        }
      },

      fetchModelAnalytics: async () => {
        set({ isLoadingModels: true, modelsError: null });
        try {
          const analytics = await adminService.getModelAnalytics();
          set({ modelAnalytics: analytics, isLoadingModels: false });
        } catch (error) {
          console.error('Failed to fetch model analytics:', error);
          set({
            modelsError: error instanceof Error ? error.message : 'Failed to fetch model analytics',
            isLoadingModels: false,
          });
        }
      },

      fetchChatAnalytics: async () => {
        set({ isLoadingChats: true, chatsError: null });
        try {
          const analytics = await adminService.getChatAnalytics();
          set({ chatAnalytics: analytics, isLoadingChats: false });
        } catch (error) {
          console.error('Failed to fetch chat analytics:', error);
          set({
            chatsError: error instanceof Error ? error.message : 'Failed to fetch chat analytics',
            isLoadingChats: false,
          });
        }
      },

      fetchAgentAnalytics: async () => {
        set({ isLoadingAgents: true, agentsError: null });
        try {
          const analytics = await adminService.getAgentAnalytics();
          set({ agentAnalytics: analytics, isLoadingAgents: false });
        } catch (error) {
          console.error('Failed to fetch agent analytics:', error);
          set({
            agentsError: error instanceof Error ? error.message : 'Failed to fetch agent analytics',
            isLoadingAgents: false,
          });
        }
      },

      refreshAllAnalytics: async () => {
        // Fetch all analytics in parallel
        await Promise.allSettled([
          adminService.getOverviewStats().then(stats => set({ overviewStats: stats })),
          adminService
            .getProviderAnalytics()
            .then(analytics => set({ providerAnalytics: analytics })),
          adminService.getModelAnalytics().then(analytics => set({ modelAnalytics: analytics })),
          adminService.getChatAnalytics().then(analytics => set({ chatAnalytics: analytics })),
          adminService.getAgentAnalytics().then(analytics => set({ agentAnalytics: analytics })),
        ]);
      },
    }),
    {
      name: 'admin-store',
    }
  )
);
