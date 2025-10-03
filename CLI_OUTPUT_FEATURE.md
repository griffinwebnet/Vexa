# CLI Output Streaming Feature

## Overview

This feature adds real-time CLI output streaming to the domain provisioning process, allowing users to see exactly what's happening during domain setup and copy the output for debugging purposes.

## Features

- **Real-time CLI output**: See samba-tool commands and their output in real-time
- **Color-coded output**: Different colors for stdout, stderr, and errors
- **Copy to clipboard**: Copy the entire CLI output for logging/debugging
- **Auto-scrolling**: Terminal output automatically scrolls to show the latest messages
- **Continue button**: Option to proceed to dashboard after viewing output

## API Changes

### New Endpoint

- `POST /api/v1/domain/provision-with-output` - Streams CLI output during domain provisioning

### Server-Sent Events (SSE)

The endpoint uses Server-Sent Events to stream output in real-time:

```
data: {"type": "output", "content": "Starting domain provisioning...", "timestamp": 1234567890}
data: {"type": "output", "content": "STDOUT: Provisioning domain...", "timestamp": 1234567890}
data: {"type": "complete", "content": "Domain provisioning completed", "timestamp": 1234567890}
```

## Frontend Changes

### SetupWizard Component

- Added CLI output display with terminal-like styling
- Real-time output streaming using fetch API with ReadableStream
- Copy to clipboard functionality
- Color-coded output (green for stdout, yellow for stderr, red for errors)
- Auto-scrolling terminal output
- Continue button to proceed after viewing output

## Usage

1. Start domain provisioning from the Setup Wizard
2. CLI output will appear in real-time in the terminal window
3. Use the "Copy" button to copy all output to clipboard
4. After completion, click "Continue to Dashboard" to proceed

## Technical Implementation

### Backend

- `DomainHandler.ProvisionDomainWithOutput()` - New handler for streaming endpoint
- `DomainService.ProvisionDomainWithOutput()` - Service method with output streaming
- `SambaTool.DomainProvisionWithOutput()` - Exec tool with real-time output capture

### Frontend

- Uses fetch API with ReadableStream for SSE
- Real-time output parsing and display
- Clipboard API for copying output
- Auto-scrolling with useRef

## Benefits

1. **Debugging**: See exactly what's happening during domain provisioning
2. **Transparency**: No more silent failures - all output is visible
3. **Logging**: Copy output for support tickets or documentation
4. **User Experience**: Better feedback during long-running operations

## Error Handling

- Network errors are caught and displayed
- SSE parsing errors are handled gracefully
- Failed commands show error output in red
- Continue button allows proceeding even if some steps failed

## Future Enhancements

- Add progress indicators for long-running operations
- Save output to server logs
- Add filtering options (show only errors, etc.)
- Export output to file
