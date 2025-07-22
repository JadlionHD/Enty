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
  })

  const fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  tab.fitAddon = fitAddon

  // Handle user input
  terminal.onData((data) => {
    if (tab.isRunning) {
      WriteToTerminal(tab.id, data).catch((error) => {
        console.error('Error writing to terminal:', error)
        terminal.writeln(`\r\nError: ${error}`)
      })
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

  // Wait for DOM update then initialize
  await nextTick()
  await initializeTab(tab)
}

const switchToTab = async (tabId: string) => {
  if (activeTabId.value === tabId) return

  activeTabId.value = tabId
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

  // Create initial tab
  if (availableTerminalTypes.value.length > 0) {
    await addTab(availableTerminalTypes.value[0])
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
  height: 100%;
  min-height: 400px;
}

.terminal-content {
  flex: 1;
  position: relative;
}

.terminal-tab {
  padding: 8px;
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

/* Ensure xterm.js terminal fits properly */
:deep(.xterm) {
  height: 100% !important;
}

:deep(.xterm-viewport) {
  height: 100% !important;
}

/* Add tab styling */
.terminal-header {
  border-bottom: 1px solid #374151;
}

/* Dropdown menu styling */
.relative {
  position: relative;
}

/* Close button styling */
button {
  transition: background-color 0.2s ease;
}
</style>
