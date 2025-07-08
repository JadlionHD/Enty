# Enty

A Moe local development app :>

## Features

- **Cross-platform support**: Windows, Linux, and macOS

- **Package management**: Easy installation and management of software packages

- **Modern UI**: Built with Vue.js and TypeScript for a responsive user experience

- **Fast performance**: Powered by Go backend with efficient package handling

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.23 or higher)
- [Node.js](https://nodejs.org/) (version 22 or higher)
- [Wails](https://wails.io/docs/gettingstarted/installation)

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd Enty
   ```

2. Install dependencies:
   ```bash
   wails doctor
   ```

## Development

### Live Development Mode

To run in live development mode with hot reload:

```bash
make dev
# or
wails dev
```

This will start:

- A Vite development server for the frontend
- A development server on http://localhost:34115 for browser testing
- Hot reload for both frontend and backend changes

### Available Make Commands

```bash
make dev          # Start development server
make build-app    # Build production application
make clean        # Clean build artifacts
```

## Building

To build a redistributable production package:

```bash
make build-app
```

## License

This project is licensed under the GPL-3 License - see the LICENSE file for details.
