<template>
  <div class="terminal-container w-full h-full bg-black">
    <div class="terminal-header bg-gray-800 text-white px-4 py-2 flex justify-between items-center">
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
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-600 hover:bg-gray-500 text-gray-200'
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
            class="px-2 py-1 text-xs bg-green-600 hover:bg-green-700 rounded"
          >
            + Add Tab
          </button>
          <div
            v-if="showAddTabMenu"
            class="absolute top-full left-0 mt-1 bg-gray-700 border border-gray-600 rounded shadow-lg z-10"
          >
            <button
              v-for="terminalType in availableTerminalTypes"
              :key="terminalType"
              @click="addTab(terminalType)"
              class="block w-full px-4 py-2 text-left text-xs hover:bg-gray-600 text-white"
            >
              {{ terminalType.charAt(0).toUpperCase() + terminalType.slice(1) }}
            </button>
          </div>
        </div>
      </div>

      <div class="flex gap-2">
        <button 
          @click="clearActiveTerminal"
          class="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 rounded"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Terminal Content Area -->
    <div class="terminal-content flex-1">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        :ref="(el: any) => setTerminalRef(tab.id, el as HTMLElement)"
        :class="['terminal-tab w-full h-full', { 'hidden': activeTabId !== tab.id }]"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { 
  CreateTerminalSession, 
  WriteToTerminal, 
  CloseTerminalSession, 
  ResizeTerminalSession, 
  IsTerminalSessionRunning,
  GetAvailableTerminalTypes,
  ListTerminalSessions
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface TerminalTab {
  id: string
  name: string
  type: string
  terminal: Terminal | null
  fitAddon: FitAddon | null
  isRunning: boolean
}

const tabs = ref<TerminalTab[]>([])
const activeTabId = ref<string>('')
const showAddTabMenu = ref(false)
const availableTerminalTypes = ref<string[]>([])
const terminalRefs = reactive<Map<string, HTMLElement>>(new Map())

// Persistence key for localStorage
const TERMINAL_STATE_KEY = 'enty-terminal-state'

// Save terminal state to localStorage
const saveTerminalState = () => {
  const state = {
    tabs: tabs.value.map(tab => ({
      id: tab.id,
      name: tab.name,
      type: tab.type,
      // Note: Don't save terminal instances or running state
    })),
    activeTabId: activeTabId.value,
  }
  localStorage.setItem(TERMINAL_STATE_KEY, JSON.stringify(state))
}

// Load terminal state from localStorage
const loadTerminalState = () => {
  const saved = localStorage.getItem(TERMINAL_STATE_KEY)
  if (saved) {
    try {
      const state = JSON.parse(saved)
      return state
    } catch (error) {
      console.warn('Failed to parse saved terminal state:', error)
    }
  }
  return null
}

// Generate unique session ID
const generateSessionId = () => `terminal-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`

const setTerminalRef = (tabId: string, el: HTMLElement | null) => {
  if (el) {
    terminalRefs.set(tabId, el)
  } else {
    terminalRefs.delete(tabId)
  }
}

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
    cols: 80,
    rows: 24,
    // Performance optimizations
    scrollback: 1000,      // Limit scrollback for better performance
    fastScrollModifier: 'alt',
    fastScrollSensitivity: 5,
    scrollSensitivity: 3,
  })

  const fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  tab.fitAddon = fitAddon

  // Handle user input with batching for better performance
  let inputBuffer = ''
  let inputTimeout: NodeJS.Timeout | null = null

  terminal.onData((data) => {
    if (tab.isRunning) {
      // Batch input for better performance (especially for rapid typing)
      inputBuffer += data
      
      if (inputTimeout) {
        clearTimeout(inputTimeout)
      }
      
      inputTimeout = setTimeout(() => {
        WriteToTerminal(tab.id, inputBuffer).catch((error) => {
          console.error('Error writing to terminal:', error)
          terminal.writeln(`\r\nError: ${error}`)
        })
        inputBuffer = ''
        inputTimeout = null
      }, 1) // Small delay for batching
    }
  })

  // Handle resize events
  terminal.onResize(({ cols, rows }) => {
    if (tab.isRunning) {
      ResizeTerminalSession(tab.id, cols, rows).catch((error: any) => {
        console.error('Error resizing terminal:', error)
      })
    }
  })

  return terminal
}

const initializeTab = async (tab: TerminalTab) => {
  const element = terminalRefs.get(tab.id)
  if (!element || tab.terminal) return

  tab.terminal = createTerminalInstance(tab)
  tab.terminal.open(element)

  await nextTick()
  if (tab.fitAddon) {
    tab.fitAddon.fit()
  }

  // Automatically start terminal session
  try {
    await CreateTerminalSession(tab.id, tab.type)
    tab.isRunning = true
    tab.terminal.focus()
  } catch (error) {
    console.error('Error starting terminal session:', error)
    tab.terminal.writeln(`Error starting ${tab.type} session: ${error}`)
  }
}

const addTab = async (terminalType: string) => {
  const sessionId = generateSessionId()
  const tab: TerminalTab = {
    id: sessionId,
    name: `${terminalType.charAt(0).toUpperCase() + terminalType.slice(1)}`,
    type: terminalType,
    terminal: null,
    fitAddon: null,
    isRunning: false
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
  showAddTabMenu.value = false
  
  // Save state to localStorage
  saveTerminalState()

  // Wait for DOM update then initialize
  await nextTick()
  await initializeTab(tab)
}

const switchToTab = async (tabId: string) => {
  if (activeTabId.value === tabId) return

  activeTabId.value = tabId
  saveTerminalState() // Save active tab change
  await nextTick()

  const tab = tabs.value.find(t => t.id === tabId)
  if (!tab) return

  if (!tab.terminal) {
    await initializeTab(tab)
  } else {
    // Re-fit terminal when switching tabs
    if (tab.fitAddon) {
      tab.fitAddon.fit()
    }
    tab.terminal.focus()
  }
}

const closeTab = async (tabId: string) => {
  const tabIndex = tabs.value.findIndex(t => t.id === tabId)
  if (tabIndex === -1) return

  const tab = tabs.value[tabIndex]
  
  // Close terminal session
  try {
    await CloseTerminalSession(tabId)
  } catch (error) {
    console.error('Error closing terminal session:', error)
  }

  // Dispose terminal instance
  if (tab.terminal) {
    tab.terminal.dispose()
  }

  // Remove tab
  tabs.value.splice(tabIndex, 1)
  terminalRefs.delete(tabId)
  
  // Save state after tab removal
  saveTerminalState()

  // Switch to another tab if this was active
  if (activeTabId.value === tabId) {
    if (tabs.value.length > 0) {
      const newActiveIndex = Math.max(0, tabIndex - 1)
      switchToTab(tabs.value[newActiveIndex].id)
    } else {
      // Create a default tab if no tabs remain
      const types = await GetAvailableTerminalTypes()
      if (types.length > 0) {
        addTab(types[0])
      }
    }
  }
}

const clearActiveTerminal = () => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.terminal) {
    activeTab.terminal.clear()
  }
}

const handleTerminalData = (event: { sessionID: string, data: string }) => {
  const tab = tabs.value.find(t => t.id === event.sessionID)
  if (tab?.terminal) {
    // Optimized terminal output writing
    tab.terminal.write(event.data)
  }
}

const handleTerminalExit = (event: { sessionID: string, message: string }) => {
  const tab = tabs.value.find(t => t.id === event.sessionID)
  if (tab) {
    tab.isRunning = false
    if (tab.terminal) {
      tab.terminal.writeln(`\r\n${event.message}`)
      tab.terminal.writeln('Terminal session ended due to timeout or process exit.')
    }
  }
}

const handleResize = () => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.fitAddon) {
    activeTab.fitAddon.fit()
  }
}

// Close dropdown when clicking outside
const handleClickOutside = (event: Event) => {
  const target = event.target as Element
  if (!target.closest('.relative')) {
    showAddTabMenu.value = false
  }
}

onMounted(async () => {
  // Load available terminal types
  try {
    availableTerminalTypes.value = await GetAvailableTerminalTypes()
  } catch (error) {
    console.error('Error loading terminal types:', error)
    availableTerminalTypes.value = ['bash'] // fallback
  }

  // Load saved terminal state
  const savedState = loadTerminalState()
  
  if (savedState && savedState.tabs && savedState.tabs.length > 0) {
    // Restore saved tabs
    for (const savedTab of savedState.tabs) {
      const tab: TerminalTab = {
        id: savedTab.id,
        name: savedTab.name,
        type: savedTab.type,
        terminal: null,
        fitAddon: null,
        isRunning: false
      }
      tabs.value.push(tab)
    }
    
    // Set active tab
    if (savedState.activeTabId && tabs.value.find(t => t.id === savedState.activeTabId)) {
      activeTabId.value = savedState.activeTabId
    } else {
      activeTabId.value = tabs.value[0].id
    }
    
    // Initialize active tab
    await nextTick()
    const activeTab = tabs.value.find(t => t.id === activeTabId.value)
    if (activeTab) {
      await initializeTab(activeTab)
    }
  } else {
    // Create initial tab if no saved state
    if (availableTerminalTypes.value.length > 0) {
      await addTab(availableTerminalTypes.value[0])
    }
  }

  // Listen for terminal events from backend
  EventsOn('terminal:data', handleTerminalData)
  EventsOn('terminal:exit', handleTerminalExit)

  // Handle window resize
  window.addEventListener('resize', handleResize)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  // Clean up all terminals
  tabs.value.forEach(tab => {
    if (tab.terminal) {
      tab.terminal.dispose()
    }
    if (tab.isRunning) {
      CloseTerminalSession(tab.id).catch(console.error)
    }
  })
  
  // Remove event listeners
  EventsOff('terminal:data')
  EventsOff('terminal:exit')
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 80vh; /* 80% of viewport height as requested */
  min-height: 400px;
  max-height: calc(100vh - 100px); /* Leave some space for other UI elements */
}

.terminal-content {
  flex: 1;
  position: relative;
  overflow: hidden; /* Prevent scrollbars outside terminals */
}

.terminal-tab {
  padding: 8px;
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

/* Ensure xterm.js terminal fits properly and performs well */
:deep(.xterm) {
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  height: 100% !important;
  width: 100% !important;
}

/* Optimize xterm.js performance */
:deep(.xterm-screen) {
  /* Enable hardware acceleration for better performance */
  transform: translateZ(0);
  will-change: transform;
}

:deep(.xterm .xterm-cursor-layer) {
  /* Smoother cursor animations */
  transition: opacity 0.1s ease;
}

/* Add tab styling with better UX */
.terminal-header {
  border-bottom: 1px solid #374151;
  background: linear-gradient(135deg, #1f2937 0%, #111827 100%);
  min-height: 48px; /* Ensure adequate tab height */
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
    height: 75vh; /* Slightly less on smaller screens */
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
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-track {
  background: #1f2937;
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 4px;
}

:deep(.xterm .xterm-viewport)::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>
