<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, reactive, watch } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { ClipboardAddon } from '@xterm/addon-clipboard';
import '@xterm/xterm/css/xterm.css';
import { useWindowSize } from '@vueuse/core';
import {
  CreateTerminalSession,
  WriteToTerminal,
  CloseTerminalSession,
  ResizeTerminalSession,
  GetAvailableTerminalTypes,
} from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

interface TerminalTab {
  id: string;
  name: string;
  type: string;
  terminal: Terminal | null;
  fitAddon: FitAddon | null;
  isRunning: boolean;
  isPaused: boolean; // New: Track if terminal is paused to save resources
  lastActiveTime: number; // New: Track last activity for resource optimization
}

const tabs = ref<TerminalTab[]>([]);
const activeTabId = ref<string>('');
const showAddTabMenu = ref(false);
const availableTerminalTypes = ref<string[]>([]);
const terminalRefs = reactive<Map<string, HTMLElement>>(new Map());
const isComponentVisible = ref(true); // Track component visibility for resource optimization
const resourceOptimizationTimer = ref<number | null>(null); // Timer for pausing inactive terminals
const resourceOptimizationInterval = ref<number | null>(null); // Interval for periodic resource optimization
const contextMenuVisible = ref(false); // Track context menu visibility
const contextMenuPosition = ref({ x: 0, y: 0 }); // Context menu position
const documentClickHandler = ref<((event: Event) => void) | null>(null); // Store document click handler ref
const visibilityChangeHandler = ref<(() => void) | null>(null); // Store visibility change handler ref

// Use VueUse for window size tracking
const { width, height } = useWindowSize();

// Persistence key for localStorage
const TERMINAL_STATE_KEY = 'enty-terminal-state';

// Save terminal state to localStorage
const saveTerminalState = () => {
  const state = {
    tabs: tabs.value.map((tab) => ({
      id: tab.id,
      name: tab.name,
      type: tab.type,
      // Note: Don't save terminal instances or running state
    })),
    activeTabId: activeTabId.value,
  };
  localStorage.setItem(TERMINAL_STATE_KEY, JSON.stringify(state));
};

// Load terminal state from localStorage
const loadTerminalState = () => {
  const saved = localStorage.getItem(TERMINAL_STATE_KEY);
  if (saved) {
    try {
      const state = JSON.parse(saved);
      return state;
    } catch (error) {
      console.warn('Failed to parse saved terminal state:', error);
    }
  }
  return null;
};

// Generate unique session ID
const generateSessionId = () => `terminal-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

const setTerminalRef = (tabId: string, el: HTMLElement | null) => {
  if (el) {
    terminalRefs.set(tabId, el);
  } else {
    terminalRefs.delete(tabId);
  }
};

const createTerminalInstance = (tab: TerminalTab): Terminal => {
  const terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#000000',
      foreground: '#ffffff',
      cursor: '#ffffff',
    },
    cols: 200, // Increased columns for full width
    rows: 24,
    scrollback: 500,
    fastScrollModifier: 'alt',
    fastScrollSensitivity: 5,
    scrollSensitivity: 3,
    windowsMode: false, // Disable Windows-specific rendering
    macOptionIsMeta: false, // Optimize for performance
    disableStdin: false, // Keep input enabled
    allowProposedApi: true, // Enable performance API
  });

  const fitAddon = new FitAddon();
  const clipboardAddon = new ClipboardAddon();

  terminal.loadAddon(fitAddon);
  terminal.loadAddon(clipboardAddon);
  tab.fitAddon = fitAddon;

  // Add right-click context menu for copy/paste
  terminal.attachCustomKeyEventHandler((event) => {
    // Handle Ctrl+C and Ctrl+V
    if (event.ctrlKey && event.key === 'c' && terminal.hasSelection()) {
      navigator.clipboard.writeText(terminal.getSelection());
      return false;
    }
    if (event.ctrlKey && event.key === 'v') {
      navigator.clipboard.readText().then((text) => {
        WriteToTerminal(tab.id, text).catch((error) => {
          console.error('Error writing to terminal:', error);
        });
      });
      return false;
    }
    return true;
  });

  // Optimized input handling with reduced debouncing for better responsiveness
  let inputBuffer = '';
  let inputTimeout: NodeJS.Timeout | null = null;

  terminal.onData((data) => {
    if (!tab.isRunning) return;

    // For single characters, send immediately to reduce latency
    if (data.length === 1 && inputBuffer === '') {
      WriteToTerminal(tab.id, data).catch((error) => {
        console.error('Error writing to terminal:', error);
        terminal.writeln(`\r\nError: ${error}`);
      });
      return;
    }

    // Batch only for multi-character input (paste operations)
    inputBuffer += data;

    if (inputTimeout) {
      clearTimeout(inputTimeout);
    }

    // Immediate send for better responsiveness
    inputTimeout = setTimeout(() => {
      WriteToTerminal(tab.id, inputBuffer).catch((error) => {
        console.error('Error writing to terminal:', error);
        terminal.writeln(`\r\nError: ${error}`);
      });
      inputBuffer = '';
      inputTimeout = null;
    }, 0); // Immediate execution in next tick
  });

  // Throttled resize handling to prevent excessive calls
  let resizeTimeout: NodeJS.Timeout | null = null;
  terminal.onResize(({ cols, rows }) => {
    if (!tab.isRunning) return;

    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }

    resizeTimeout = setTimeout(() => {
      ResizeTerminalSession(tab.id, cols, rows).catch((error: any) => {
        console.error('Error resizing terminal:', error);
      });
      resizeTimeout = null;
    }, 50); // Throttle resize events
  });

  return terminal;
};

const initializeTab = async (tab: TerminalTab) => {
  const element = terminalRefs.get(tab.id);
  if (!element || tab.terminal) return;

  tab.terminal = createTerminalInstance(tab);
  tab.terminal.open(element);

  await nextTick();

  // Initial fitting
  if (tab.fitAddon && tab.terminal) {
    const containerWidth = element.clientWidth;
    const containerHeight = element.clientHeight;

    // Calculate optimal dimensions
    const fontSize = 14;
    const charWidth = fontSize * 0.6;
    const charHeight = fontSize * 1.2;

    // Calculate cols and rows
    const cols = Math.max(10, Math.floor(containerWidth / charWidth));
    const rows = Math.max(10, Math.floor(containerHeight / charHeight));

    // Set initial dimensions
    tab.terminal.resize(cols, rows);
    tab.fitAddon.fit();

    // Set minimum and maximum row limits
    const finalRows = Math.min(Math.max(rows, 10), 50);

    // Force terminal to use calculated dimensions
    tab.terminal.resize(80, finalRows);
    await nextTick();
    tab.fitAddon.fit();
  }

  // Automatically start terminal session
  try {
    await CreateTerminalSession(tab.id, tab.type);
    tab.isRunning = true;

    // Focus terminal after short delay to ensure proper rendering
    setTimeout(() => {
      if (tab.terminal && activeTabId.value === tab.id) {
        tab.terminal.focus();
      }
    }, 100);
  } catch (error) {
    console.error('Error starting terminal session:', error);
    if (tab.terminal) {
      tab.terminal.writeln(`Error starting ${tab.type} session: ${error}`);
    }
  }
};

const addTab = async (terminalType: string) => {
  const sessionId = generateSessionId();
  const tab: TerminalTab = {
    id: sessionId,
    name: `${terminalType.charAt(0).toUpperCase() + terminalType.slice(1)}`,
    type: terminalType,
    terminal: null,
    fitAddon: null,
    isRunning: false,
    isPaused: false,
    lastActiveTime: Date.now(),
  };

  tabs.value.push(tab);
  activeTabId.value = tab.id;
  showAddTabMenu.value = false;

  // Save state to localStorage
  saveTerminalState();

  // Wait for DOM update then initialize
  await nextTick();
  await initializeTab(tab);
};

const switchToTab = async (tabId: string) => {
  if (activeTabId.value === tabId) return;

  // Update last active time for previous tab
  const prevTab = tabs.value.find((t) => t.id === activeTabId.value);
  if (prevTab) {
    prevTab.lastActiveTime = Date.now();
  }

  // Set new active tab
  activeTabId.value = tabId;
  saveTerminalState(); // Save active tab change

  // Find the target tab
  const tab = tabs.value.find((t) => t.id === tabId);
  if (!tab) return;

  // Update current tab's active time
  tab.lastActiveTime = Date.now();

  // Handle tab initialization or resumption
  if (!tab.terminal || tab.isPaused) {
    await nextTick();
    tab.isPaused ? await resumeTerminal(tab) : await initializeTab(tab);
  } else {
    // Optimize terminal fitting with a single requestAnimationFrame
    requestAnimationFrame(() => {
      if (tab.fitAddon && tab.terminal) {
        tab.fitAddon.fit();
        tab.terminal.focus();
      }
    });
  }
};

const closeTab = async (tabId: string) => {
  const tabIndex = tabs.value.findIndex((t) => t.id === tabId);
  if (tabIndex === -1) return;

  const tab = tabs.value[tabIndex];
  const wasActive = activeTabId.value === tabId;

  // Dispose terminal instance first to free resources immediately
  if (tab.terminal) {
    tab.terminal.dispose();
    tab.terminal = null;
  }

  // Close terminal session in parallel with UI updates
  const closeSessionPromise = CloseTerminalSession(tabId).catch((error) =>
    console.error('Error closing terminal session:', error),
  );

  // Remove tab from array and refs map
  tabs.value.splice(tabIndex, 1);
  terminalRefs.delete(tabId);

  // Save state after tab removal
  saveTerminalState();

  // Handle tab switching if needed
  if (wasActive) {
    if (tabs.value.length > 0) {
      // Prefer the tab to the left of the closed one
      const newActiveIndex = Math.min(tabIndex, tabs.value.length - 1);
      switchToTab(tabs.value[newActiveIndex].id);
    } else {
      // Create a default tab if no tabs remain, using cached types if available
      const types =
        availableTerminalTypes.value.length > 0
          ? availableTerminalTypes.value
          : await GetAvailableTerminalTypes();

      if (types.length > 0) {
        addTab(types[0]);
      }
    }
  }

  // Ensure session is closed before returning
  await closeSessionPromise;
};

// Handle context menu for copy/paste
const handleContextMenu = (event: MouseEvent, tabId: string) => {
  event.preventDefault();
  const tab = tabs.value.find((t) => t.id === tabId);
  if (tab?.terminal) {
    contextMenuPosition.value = { x: event.clientX, y: event.clientY };
    contextMenuVisible.value = true;
  }
};

const copyToClipboard = () => {
  const activeTab = tabs.value.find((t) => t.id === activeTabId.value);
  if (activeTab?.terminal && activeTab.terminal.hasSelection()) {
    navigator.clipboard.writeText(activeTab.terminal.getSelection());
  }
  hideContextMenu();
};

const pasteFromClipboard = () => {
  const activeTab = tabs.value.find((t) => t.id === activeTabId.value);
  if (activeTab?.terminal && activeTab.isRunning) {
    navigator.clipboard.readText().then((text) => {
      WriteToTerminal(activeTab.id, text).catch((error) => {
        console.error('Error writing to terminal:', error);
      });
    });
  }
  hideContextMenu();
};

const hideContextMenu = () => {
  contextMenuVisible.value = false;
};

const clearActiveTerminal = () => {
  const activeTab = tabs.value.find((t) => t.id === activeTabId.value);
  if (activeTab?.terminal) {
    activeTab.terminal.clear();
  }
};

// Resource optimization functions
const pauseInactiveTerminals = () => {
  const now = Date.now();
  const INACTIVE_THRESHOLD = 5 * 60 * 1000; // 5 minutes of inactivity

  tabs.value.forEach((tab) => {
    if (tab.id !== activeTabId.value && !tab.isPaused && tab.terminal) {
      if (now - tab.lastActiveTime > INACTIVE_THRESHOLD) {
        // Pause terminal to save resources
        tab.isPaused = true;
        // Dispose of the terminal instance but keep the session
        if (tab.terminal) {
          tab.terminal.dispose();
          tab.terminal = null;
        }
        console.log(`Paused inactive terminal: ${tab.name}`);
      }
    }
  });
};

const resumeTerminal = async (tab: TerminalTab) => {
  if (tab.isPaused && !tab.terminal) {
    tab.isPaused = false;
    tab.lastActiveTime = Date.now();
    // Recreate terminal instance
    await initializeTab(tab);
    console.log(`Resumed terminal: ${tab.name}`);
  }
};

// Optimized visibility handlers using Vue lifecycle instead of direct DOM manipulation
const handleVisibilityChange = () => {
  isComponentVisible.value = !document.hidden;

  if (!isComponentVisible.value) {
    // Component not visible, start resource optimization timer
    resourceOptimizationTimer.value = window.setTimeout(() => {
      pauseInactiveTerminals();
    }, 30000); // Pause after 30 seconds of invisibility
  } else {
    // Component visible, cancel optimization timer
    if (resourceOptimizationTimer.value) {
      clearTimeout(resourceOptimizationTimer.value);
      resourceOptimizationTimer.value = null;
    }
  }
};

// Close dropdown when clicking outside using proper event handler refs
const handleClickOutside = (event: Event) => {
  const target = event.target as Element;
  if (!target.closest('.relative')) {
    showAddTabMenu.value = false;
  }
  // Also hide context menu when clicking outside
  if (contextMenuVisible.value && !target.closest('.context-menu')) {
    hideContextMenu();
  }
};

const handleTerminalData = (event: { sessionID: string; data: string }) => {
  const tab = tabs.value.find((t) => t.id === event.sessionID);
  if (tab?.terminal && !tab.isPaused) {
    // Only process data for active, non-paused terminals
    tab.lastActiveTime = Date.now();

    // Use requestAnimationFrame for smoother rendering
    requestAnimationFrame(() => {
      if (tab.terminal && !tab.isPaused) {
        tab.terminal.write(event.data);
      }
    });
  }
};

const handleTerminalExit = (event: { sessionID: string; message: string }) => {
  const tab = tabs.value.find((t) => t.id === event.sessionID);
  if (tab) {
    tab.isRunning = false;
    tab.lastActiveTime = Date.now();

    if (tab.terminal && !tab.isPaused) {
      requestAnimationFrame(() => {
        if (tab.terminal) {
          tab.terminal.writeln(`\r\n${event.message}`);
          tab.terminal.writeln('Terminal session ended due to timeout or process exit.');
        }
      });
    }
  }
};

// Add resize timeout ref
const resizeTimeout = ref<number | null>(null);

const handleResize = () => {
  const activeTab = tabs.value.find((t) => t.id === activeTabId.value);
  if (activeTab?.fitAddon && activeTab.terminal) {
    // Throttle resize operations to improve performance
    if (resizeTimeout.value) {
      clearTimeout(resizeTimeout.value);
    }
    resizeTimeout.value = window.setTimeout(() => {
      if (activeTab.fitAddon && activeTab.terminal) {
        try {
          // Get the container dimensions
          const container = terminalRefs.get(activeTab.id);
          if (container) {
            const containerHeight = container.clientHeight;
            const containerWidth = container.clientWidth;

            // Calculate optimal dimensions
            const fontSize = 14; // Match with terminal creation
            const charWidth = fontSize * 0.6; // Approximate width of a character
            const charHeight = fontSize * 1.2; // Approximate height of a line

            // Calculate cols and rows
            const cols = Math.max(10, Math.floor(containerWidth / charWidth));
            const rows = Math.max(10, Math.floor(containerHeight / charHeight));

            // Update terminal dimensions
            activeTab.terminal.resize(cols, rows);
            activeTab.fitAddon.fit();
          }
        } catch (error) {
          console.warn('Terminal resize failed:', error);
        }
      }
    }, 50); // Reduced throttling for better responsiveness
  }
};

// Watch for window size changes using VueUse with throttling
watch([width, height], () => {
  handleResize();
});

onMounted(async () => {
  // Load available terminal types
  try {
    availableTerminalTypes.value = await GetAvailableTerminalTypes();
  } catch (error) {
    console.error('Error loading terminal types:', error);
    availableTerminalTypes.value = ['bash']; // fallback
  }

  // Load saved terminal state
  const savedState = loadTerminalState();

  if (savedState && savedState.tabs && savedState.tabs.length > 0) {
    // Restore saved tabs
    for (const savedTab of savedState.tabs) {
      const tab: TerminalTab = {
        id: savedTab.id,
        name: savedTab.name,
        type: savedTab.type,
        terminal: null,
        fitAddon: null,
        isRunning: false,
        isPaused: false,
        lastActiveTime: Date.now(),
      };
      tabs.value.push(tab);
    }

    // Set active tab
    if (savedState.activeTabId && tabs.value.find((t) => t.id === savedState.activeTabId)) {
      activeTabId.value = savedState.activeTabId;
    } else {
      activeTabId.value = tabs.value[0].id;
    }

    // Initialize active tab
    await nextTick();
    const activeTab = tabs.value.find((t) => t.id === activeTabId.value);
    if (activeTab) {
      await initializeTab(activeTab);
    }
  } else {
    // Create initial tab if no saved state
    if (availableTerminalTypes.value.length > 0) {
      await addTab(availableTerminalTypes.value[0]);
    }
  }

  // Listen for terminal events from backend
  EventsOn('terminal:data', handleTerminalData);
  EventsOn('terminal:exit', handleTerminalExit);

  // Handle click outside and visibility changes (resize is handled by VueUse watcher)
  // Setup event handlers using refs to avoid direct DOM manipulation
  documentClickHandler.value = handleClickOutside;
  visibilityChangeHandler.value = handleVisibilityChange;

  // Handle window resize and visibility changes using proper cleanup
  window.addEventListener('resize', handleResize);
  document.addEventListener('click', documentClickHandler.value);
  document.addEventListener('visibilitychange', visibilityChangeHandler.value);

  // Start periodic resource optimization (every 2 minutes)
  resourceOptimizationInterval.value = window.setInterval(
    () => {
      if (isComponentVisible.value) {
        pauseInactiveTerminals();
      }
    },
    2 * 60 * 1000,
  );
});

onUnmounted(() => {
  // Clean up resize timeout
  if (resizeTimeout.value) {
    clearTimeout(resizeTimeout.value);
  }

  // Clean up resource optimization timers
  if (resourceOptimizationTimer.value) {
    clearTimeout(resourceOptimizationTimer.value);
  }

  // Clean up resource optimization interval
  if (resourceOptimizationInterval.value) {
    clearInterval(resourceOptimizationInterval.value);
    resourceOptimizationInterval.value = null;
  }

  // Clean up all terminals and dispose properly
  tabs.value.forEach((tab) => {
    if (tab.terminal) {
      tab.terminal.dispose();
    }
    if (tab.isRunning) {
      CloseTerminalSession(tab.id).catch(console.error);
    }
  });

  // Remove event listeners using refs for proper cleanup
  EventsOff('terminal:data');
  EventsOff('terminal:exit');

  if (documentClickHandler.value) {
    document.removeEventListener('click', documentClickHandler.value);
    documentClickHandler.value = null;
  }

  if (visibilityChangeHandler.value) {
    document.removeEventListener('visibilitychange', visibilityChangeHandler.value);
    visibilityChangeHandler.value = null;
  }
});
</script>

<template>
  <div class="terminal-container w-full h-full">
    <div class="terminal-header text-white px-4 py-2 flex justify-between items-center">
      <div class="flex items-center gap-2">
        <!-- Tab Navigation -->
        <div class="flex gap-1">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="switchToTab(tab.id)"
            :class="[
              'px-3 py-1 text-xs rounded flex items-center gap-2',
              activeTabId === tab.id
                ? 'bg-primary-600 text-white'
                : 'bg-neutral-600 hover:bg-neutral-500 text-neutral-200',
            ]"
          >
            <span>{{ tab.name }}</span>
            <button
              @click.stop="closeTab(tab.id)"
              class="hover:bg-red-500 rounded px-1"
              v-if="tabs.length > 1"
            >
              ×
            </button>
          </button>
        </div>

        <!-- Add Tab Dropdown -->
        <div class="relative">
          <button
            @click="showAddTabMenu = !showAddTabMenu"
            class="px-2 py-1 text-xs bg-primary-600 hover:bg-primary-700 rounded"
          >
            + Add Tab
          </button>
          <div
            v-if="showAddTabMenu"
            class="absolute top-full left-0 mt-1 bg-neutral-700 border border-neutral-600 rounded shadow-lg z-10"
          >
            <button
              v-for="terminalType in availableTerminalTypes"
              :key="terminalType"
              @click="addTab(terminalType)"
              class="block w-full px-4 py-2 text-left text-xs hover:bg-neutral-600 text-white"
            >
              {{ terminalType.charAt(0).toUpperCase() + terminalType.slice(1) }}
            </button>
          </div>
        </div>
      </div>

      <div class="flex gap-2">
        <button
          @click="clearActiveTerminal"
          class="px-3 py-1 text-xs bg-secondary-600 hover:bg-secondary-700 rounded"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Terminal Content Area -->
    <div class="terminal-content flex-1 h-full">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        :ref="(el: any) => setTerminalRef(tab.id, el as HTMLElement)"
        :class="['terminal-tab', { hidden: activeTabId !== tab.id }]"
        @contextmenu="handleContextMenu($event, tab.id)"
      ></div>
    </div>

    <!-- Context Menu for Copy/Paste -->
    <div
      v-if="contextMenuVisible"
      :style="{
        position: 'fixed',
        left: contextMenuPosition.x + 'px',
        top: contextMenuPosition.y + 'px',
        zIndex: 1000,
      }"
      class="bg-neutral-700 border border-neutral-600 rounded shadow-lg py-1 min-w-[120px]"
    >
      <button
        @click="copyToClipboard"
        class="block w-full px-4 py-2 text-left text-sm text-white hover:bg-gray-600 cursor-pointer"
      >
        Copy
      </button>
      <button
        @click="pasteFromClipboard"
        class="block w-full px-4 py-2 text-left text-sm text-white hover:bg-gray-600 cursor-pointer"
      >
        Paste
      </button>
    </div>
  </div>
</template>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 80vh;
  /* 80% of viewport height as requested */
  min-height: 400px;
  max-height: calc(100vh - 100px);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.terminal-content {
  flex: 1;
  position: relative;
  overflow: hidden;
  /* Prevent scrollbars outside terminals */
}

.terminal-tab {
  /* Remove padding to prevent prompt cutoff */
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* Ensure xterm.js terminal fits properly and performs well */
:deep(.xterm) {
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  height: 100% !important;
  width: 100% !important;
  overflow-y: auto !important;
  scrollbar-width: thin;
  scrollbar-color: rgba(75, 85, 99, 0.8) transparent;
}

/* Fix xterm.js helper textarea positioning */
:deep(.xterm-helper-textarea) {
  position: fixed !important;
  top: -9999px !important;
  left: -9999px !important;
  opacity: 0 !important;
  width: 0 !important;
  height: 0 !important;
  z-index: -1000 !important;
}

/* Ensure proper terminal sizing */
:deep(.xterm-screen) {
  /* Enable hardware acceleration for better performance */
  transform: translateZ(0);
  will-change: transform;
  position: relative;
}

:deep(.xterm .xterm-cursor-layer) {
  /* Smoother cursor animations */
  transition: opacity 0.1s ease;
}

/* Add tab styling with better UX */
.terminal-header {
  border-bottom: 1px solid #374151;
  background: linear-gradient(
    135deg,
    rgb(var(--color-neutral-900)) 0%,
    rgb(var(--color-neutral-950)) 100%
  );
  min-height: 48px;
  /* Ensure adequate tab height */
}

/* Dropdown menu styling */
.relative {
  position: relative;
}

/* Enhanced button styling for better UX */
button {
  transition: all 0.2s ease;
  border: none;
  cursor: pointer;
}

button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

button:active {
  transform: translateY(0);
}

/* Tab close button styling */
button:hover .close-btn {
  background-color: #ef4444;
  color: white;
}

/* Responsive design adjustments */
@media (max-height: 600px) {
  .terminal-container {
    height: 75vh;
    /* Slightly less on smaller screens */
    min-height: 300px;
  }
}

@media (max-height: 400px) {
  .terminal-container {
    height: 70vh;
    min-height: 250px;
  }

  .terminal-header {
    min-height: 40px;
  }
}

/* Performance optimizations for smooth rendering */
.terminal-container * {
  box-sizing: border-box;
}

/* Smooth tab transitions */
.terminal-tab {
  transition: opacity 0.1s ease;
}

.terminal-tab.hidden {
  opacity: 0;
  pointer-events: none;
}

/* Loading state for terminals */
.terminal-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 14px;
}

/* Better scrollbar styling for webkit browsers */
:deep(.xterm .xterm-viewport)::-webkit-scrollbar {
  width: 8px;
  background: rgba(31, 41, 55, 0.3);
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-track {
  background: rgba(31, 41, 55, 0.1);
  border-radius: 4px;
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-thumb {
  background: rgba(75, 85, 99, 0.8);
  border-radius: 4px;
  border: 1px solid rgba(55, 65, 81, 0.5);
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 114, 128, 0.9);
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-corner {
  background: transparent;
}

/* Ensure scrollbar never overlaps terminal content */
:deep(.xterm .xterm-rows) {
  margin-right: 8px;
}
</style>
