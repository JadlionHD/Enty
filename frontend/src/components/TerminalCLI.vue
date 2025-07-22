<template>
  <div class="terminal-container w-full h-full bg-black">
    <div class="terminal-header bg-gray-800 text-white px-4 py-2 flex justify-between items-center">
      <span class="font-mono text-sm">Terminal</span>
      <div class="flex gap-2">
        <button 
          @click="startTerminal" 
          :disabled="isRunning"
          class="px-3 py-1 text-xs bg-green-600 hover:bg-green-700 disabled:bg-gray-600 rounded"
        >
          {{ isRunning ? 'Running' : 'Start' }}
        </button>
        <button 
          @click="stopTerminal" 
          :disabled="!isRunning"
          class="px-3 py-1 text-xs bg-red-600 hover:bg-red-700 disabled:bg-gray-600 rounded"
        >
          Stop
        </button>
        <button 
          @click="clearTerminal"
          class="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 rounded"
        >
          Clear
        </button>
      </div>
    </div>
    <div ref="terminalRef" class="terminal-content flex-1"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { StartTerminal, WriteToTerminal, StopTerminal, ResizeTerminal, IsTerminalRunning } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const terminalRef = ref<HTMLElement>()
const isRunning = ref(false)

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null

const initializeTerminal = () => {
  if (!terminalRef.value) return

  // Initialize xterm.js
  terminal = new Terminal({
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

  // Initialize fit addon
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)

  // Open terminal in the DOM element
  terminal.open(terminalRef.value)
  
  // Fit terminal to container
  nextTick(() => {
    if (fitAddon) {
      fitAddon.fit()
    }
  })

  // Handle user input
  terminal.onData((data) => {
    if (isRunning.value) {
      WriteToTerminal(data).catch((error) => {
        console.error('Error writing to terminal:', error)
        terminal?.writeln(`\r\nError: ${error}`)
      })
    }
  })

  // Handle resize events
  terminal.onResize(({ cols, rows }) => {
    if (isRunning.value) {
      ResizeTerminal(cols, rows).catch((error) => {
        console.error('Error resizing terminal:', error)
      })
    }
  })

  // Welcome message
  terminal.writeln('Welcome to Enty Terminal')
  terminal.writeln('Click "Start" to begin a terminal session.')
  terminal.writeln('')
}

const startTerminal = async () => {
  try {
    await StartTerminal()
    isRunning.value = true
    terminal?.clear()
    terminal?.writeln('Terminal session started...')
    terminal?.focus()
  } catch (error) {
    console.error('Error starting terminal:', error)
    terminal?.writeln(`\r\nError starting terminal: ${error}`)
  }
}

const stopTerminal = async () => {
  try {
    await StopTerminal()
    isRunning.value = false
    terminal?.writeln('\r\nTerminal session stopped.')
  } catch (error) {
    console.error('Error stopping terminal:', error)
    terminal?.writeln(`\r\nError stopping terminal: ${error}`)
  }
}

const clearTerminal = () => {
  terminal?.clear()
}

const handleTerminalData = (data: string) => {
  terminal?.write(data)
}

const handleTerminalExit = (message: string) => {
  isRunning.value = false
  terminal?.writeln(`\r\n${message}`)
}

const checkTerminalStatus = async () => {
  try {
    isRunning.value = await IsTerminalRunning()
  } catch (error) {
    console.error('Error checking terminal status:', error)
  }
}

const handleResize = () => {
  if (fitAddon) {
    fitAddon.fit()
  }
}

onMounted(async () => {
  // Initialize terminal UI
  await nextTick()
  initializeTerminal()

  // Check initial terminal status
  await checkTerminalStatus()

  // Listen for terminal events from backend
  EventsOn('terminal:data', handleTerminalData)
  EventsOn('terminal:exit', handleTerminalExit)

  // Handle window resize
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  // Clean up terminal
  terminal?.dispose()
  
  // Remove event listeners
  EventsOff('terminal:data')
  EventsOff('terminal:exit')
  window.removeEventListener('resize', handleResize)
  
  // Stop terminal if running
  if (isRunning.value) {
    stopTerminal()
  }
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
  padding: 8px;
}

/* Ensure xterm.js terminal fits properly */
:deep(.xterm) {
  height: 100% !important;
}

:deep(.xterm-viewport) {
  height: 100% !important;
}
</style>
