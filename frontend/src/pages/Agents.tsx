import { useEffect, useRef, useMemo, useState } from 'react';
import { useParams, useLocation } from 'react-router-dom';
import {
  AgentSidebar,
  AgentChat,
  AgentOnboarding,
  WorkflowCanvas,
  AgentListView,
  DeployedListView,
  BlockSettingsPanel,
  MobileMenuButton,
  MobileSidebarOverlay,
  MobileTabLayout,
} from '@/components/agent-builder';
import { Snowfall } from '@/components/ui';
import { useAgentBuilderStore } from '@/store/useAgentBuilderStore';
import { useIsMobile } from '@/hooks';

export function Agents() {
  const {
    selectedBlockId,
    activeView,
    isSidebarCollapsed,
    toggleSidebarCollapsed,
    selectedAgentId,
    currentAgent,
    isLoadingAgents,
    loadAgentFromBackend,
    trackAgentAccess,
    setActiveView,
    fetchRecentAgents,
    workflow,
  } = useAgentBuilderStore();

  // Mobile detection - done once at top level
  const isMobile = useIsMobile();
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);

  // Close mobile sidebar after navigation
  const handleMobileNavigate = () => {
    setIsMobileSidebarOpen(false);
  };

  // Check if workflow is empty (only has start block or no blocks)
  const isWorkflowEmpty = useMemo(() => {
    if (!workflow || !workflow.blocks) return true;
    if (workflow.blocks.length === 0) return true;
    // Check if only block is the Start block (variable with operation 'read')
    if (workflow.blocks.length === 1) {
      const block = workflow.blocks[0];
      if (block.type === 'variable' && block.config?.operation === 'read') {
        return true;
      }
    }
    return false;
  }, [workflow]);

  // Get URL params for deep linking
  const { agentId: urlAgentId } = useParams<{ agentId?: string }>();
  const location = useLocation();
  const hasAutoOpenedRef = useRef(false);
  const hasHandledUrlRef = useRef(false);

  // Determine view type from URL path
  const isBuilderRoute = location.pathname.includes('/agents/builder/');
  const isDeployedRoute = location.pathname.includes('/agents/deployed/');

  // Handle URL-based navigation (deep linking)
  useEffect(() => {
    if (!urlAgentId || hasHandledUrlRef.current) return;

    const handleUrlNavigation = async () => {
      hasHandledUrlRef.current = true;
      hasAutoOpenedRef.current = true;

      console.log(
        '🔗 [AGENTS] Deep link detected:',
        urlAgentId,
        isBuilderRoute ? 'builder' : 'deployed'
      );

      // Load the agent from URL
      trackAgentAccess(urlAgentId);
      await loadAgentFromBackend(urlAgentId);

      // Set view based on route
      if (isDeployedRoute) {
        setActiveView('deployed');
      } else {
        setActiveView('canvas');
      }
    };

    handleUrlNavigation();
  }, [
    urlAgentId,
    isBuilderRoute,
    isDeployedRoute,
    loadAgentFromBackend,
    trackAgentAccess,
    setActiveView,
  ]);

  // Fetch recent agents and auto-open the most recent on first visit (only if no URL param)
  useEffect(() => {
    if (urlAgentId) return; // Skip if we have a URL param

    const initializeAgents = async () => {
      // Only run once on mount
      if (hasAutoOpenedRef.current) return;

      // Fetch recent agents from backend
      await fetchRecentAgents();

      // Get the latest state after fetch
      const { backendAgents: agents, selectedAgentId: currentSelected } =
        useAgentBuilderStore.getState();

      // If no agent is selected and we have agents, auto-open the most recent
      if (!currentSelected && agents.length > 0) {
        hasAutoOpenedRef.current = true;
        const mostRecent = agents[0];
        console.log('🚀 [AGENTS] Auto-opening most recent agent:', mostRecent.id, mostRecent.name);
        trackAgentAccess(mostRecent.id);
        await loadAgentFromBackend(mostRecent.id);
        setActiveView('canvas');
      } else if (currentSelected) {
        hasAutoOpenedRef.current = true;
      }
    };

    initializeAgents();
  }, [urlAgentId, fetchRecentAgents, loadAgentFromBackend, trackAgentAccess, setActiveView]);

  // Load agent from backend if selectedAgentId exists but currentAgent is null
  // This happens after refresh when synced agents aren't in localStorage
  useEffect(() => {
    if (selectedAgentId && !currentAgent && !isLoadingAgents) {
      console.log('🔄 [AGENTS] Loading agent from backend after refresh:', selectedAgentId);
      loadAgentFromBackend(selectedAgentId);
    }
  }, [selectedAgentId, currentAgent, isLoadingAgents, loadAgentFromBackend]);

  // =============================================================================
  // MOBILE LAYOUT
  // =============================================================================
  if (isMobile) {
    // Check if we're showing MobileTabLayout (canvas view with workflow)
    const showMobileTabLayout = activeView === 'canvas' && !isWorkflowEmpty && !isLoadingAgents;

    return (
      <div className="h-screen flex flex-col bg-[var(--color-bg-primary)] overflow-hidden">
        <Snowfall fadeAfter={5000} />
        {/* Mobile hamburger menu button - hidden when MobileTabLayout is shown (it has its own) */}
        {!showMobileTabLayout && (
          <MobileMenuButton
            isOpen={isMobileSidebarOpen}
            onToggle={() => setIsMobileSidebarOpen(!isMobileSidebarOpen)}
          />
        )}

        {/* Mobile sidebar overlay */}
        <MobileSidebarOverlay
          isOpen={isMobileSidebarOpen}
          onClose={() => setIsMobileSidebarOpen(false)}
        >
          <AgentSidebar isMobile onNavigate={handleMobileNavigate} />
        </MobileSidebarOverlay>

        {/* Mobile content area - pt-14 only when floating hamburger is shown */}
        <div
          className={`flex-1 flex flex-col overflow-hidden ${!showMobileTabLayout ? 'pt-14' : ''}`}
        >
          {/* My Agents list */}
          {activeView === 'my-agents' && <AgentListView className="flex-1" />}

          {/* Deployed agents list */}
          {activeView === 'deployed' && (
            <DeployedListView className="flex-1" initialAgentId={urlAgentId} />
          )}

          {/* Loading state */}
          {activeView === 'canvas' && isLoadingAgents && (
            <div className="flex-1 flex items-center justify-center">
              <div className="flex flex-col items-center gap-3">
                <div className="w-8 h-8 border-2 border-[var(--color-accent)] border-t-transparent rounded-full animate-spin" />
                <span className="text-sm text-[var(--color-text-secondary)]">
                  Loading workflow...
                </span>
              </div>
            </div>
          )}

          {/* Onboarding */}
          {(activeView === 'onboarding' ||
            (activeView === 'canvas' && isWorkflowEmpty && !isLoadingAgents)) && (
            <AgentOnboarding />
          )}

          {/* Canvas view - use MobileTabLayout instead of desktop canvas */}
          {showMobileTabLayout && (
            <MobileTabLayout
              onOpenSidebar={() => setIsMobileSidebarOpen(true)}
              onCloseSidebar={() => setIsMobileSidebarOpen(false)}
            />
          )}
        </div>
      </div>
    );
  }

  // =============================================================================
  // DESKTOP LAYOUT (unchanged)
  // =============================================================================
  return (
    <div className="h-screen flex bg-[var(--color-bg-primary)] overflow-hidden">
      <Snowfall fadeAfter={5000} />
      {/* Click-based Sidebar Toggle */}
      <div
        className="relative z-20 flex-shrink-0"
        style={{
          width: isSidebarCollapsed ? '60px' : '280px',
          transition: 'width 200ms ease-out',
        }}
      >
        <AgentSidebar isCollapsed={isSidebarCollapsed} onToggle={toggleSidebarCollapsed} />
      </div>

      {/* Main Content - View based routing */}
      {activeView === 'my-agents' && <AgentListView className="flex-1" />}

      {activeView === 'deployed' && (
        <DeployedListView className="flex-1" initialAgentId={urlAgentId} />
      )}

      {/* Show loading state when canvas view is loading */}
      {activeView === 'canvas' && isLoadingAgents && (
        <div className="flex-1 h-full w-full flex items-center justify-center bg-[var(--color-bg-primary)]">
          <div className="flex flex-col items-center gap-3">
            <div className="w-8 h-8 border-2 border-[var(--color-accent)] border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-[var(--color-text-secondary)]">Loading workflow...</span>
          </div>
        </div>
      )}

      {/* Show onboarding when explicitly set OR when canvas view with empty workflow */}
      {(activeView === 'onboarding' ||
        (activeView === 'canvas' && isWorkflowEmpty && !isLoadingAgents)) && <AgentOnboarding />}

      {activeView === 'canvas' && !isWorkflowEmpty && !isLoadingAgents && (
        <div className="flex-1 flex h-full">
          {/* Canvas - flexible */}
          <div className="h-full flex-1 min-w-0">
            <WorkflowCanvas className="h-full w-full" />
          </div>

          {/* Right Panel - 30% width with constraints */}
          <div
            className="h-full relative overflow-hidden flex-shrink-0"
            style={{
              width: '30%',
              minWidth: '320px',
              maxWidth: '500px',
              borderLeft: '1px solid var(--color-border)',
            }}
          >
            <div className="absolute inset-0 bg-[var(--color-bg-secondary)] bg-opacity-40 backdrop-blur-xl" />
            <div className="relative z-10 h-full">
              {selectedBlockId ? (
                <BlockSettingsPanel className="h-full" />
              ) : (
                <AgentChat className="h-full" />
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
