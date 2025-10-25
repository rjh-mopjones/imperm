# Testing Message Box Feature

## How to Test

### 1. Run the UI in mock mode:
```bash
cd /Users/roryhedderman/GolandProjects/imperm
./bin/imperm-ui --mock
```

### 2. Test Success Messages (Control Tab - Create Environment)

**Test Steps:**
1. Start the UI in mock mode
2. You should be on the "Control" tab (default)
3. Press Enter on "Build Environment"
4. Type a name like "test-env"
5. Press Enter

**Expected Result:**
- A **GREEN** message appears at the top saying: `✓ Started creating environment 'test-env'`
- The message stays for 3 seconds then automatically disappears
- Terraform logs appear on the right panel showing operation progress

### 3. Test Success Messages (Observe Tab - Delete Resource)

**Test Steps:**
1. Press Tab to switch to "Observe" tab
2. You should see a list of environments (from mock data)
3. Use ↑↓ or j/k to select an environment
4. Press 'x' to delete

**Expected Result:**
- A **GREEN** message appears at the top saying: `✓ Deleted environment: [name]`
- The message stays for 3 seconds then automatically disappears
- The resource list refreshes and the deleted item is removed

### 4. Test Error Messages

**To test error messages, you would need to:**
- For Control tab: Connect to a real server that returns errors (not mock mode)
- For Observe tab: Try to delete a resource when the API is down

**Expected Result:**
- A **RED** message appears at the top with the error details
- Format: `❌ Error: [error message]` or `❌ Failed to create environment 'name': [error]`
- The message stays for 3 seconds then automatically disappears

## Message Colors

- **Success messages**: Green (#46) with ✓ checkmark
- **Error messages**: Red (#196) with ❌ cross mark

## Message Duration

All messages automatically clear after **3 seconds**

## Message Locations

### Control Tab Layout:
```
┌─────────────────────────────────┬─────────────────────────────────┐
│ Actions                         │ Terraform Logs                  │
│                                 │                                 │
│ ✓ Started creating 'test-env'   │ Environment: test-env           │
│   ⬆️ MESSAGE APPEARS HERE        │ Status: RUNNING                 │
│                                 │                                 │
│ ┌─────────────────────────────┐ │ [2024-01-01 10:00:00] INFO...  │
│ │ Build Environment           │ │ [2024-01-01 10:00:01] INFO...  │
│ └─────────────────────────────┘ │ [2024-01-01 10:00:02] INFO...  │
│                                 │                                 │
│ ┌─────────────────────────────┐ │                                 │
│ │ Build with Options          │ │                                 │
│ └─────────────────────────────┘ │                                 │
└─────────────────────────────────┴─────────────────────────────────┘
```

### Observe Tab Layout:
```
┌─────────────────────────────────────────────────────────────────┐
│ Environments | Last update: 10:00:05                             │
├─────────────────────────────────┬─────────────────────────────────┤
│ Environments                    │ Details                         │
│                                 │                                 │
│ ✓ Deleted environment: dev-1    │ Selected: dev-env-2             │
│   ⬆️ MESSAGE APPEARS HERE        │                                 │
│                                 │ Name: dev-env-2                 │
│ NAME           STATUS    AGE    │ Namespace: default              │
│ dev-env-2      Running   2h     │ Status: Running                 │
│ staging-env-1  Running   1d     │ Age: 2 hours                    │
│                                 │                                 │
└─────────────────────────────────┴─────────────────────────────────┘
```

## Troubleshooting

If you don't see messages:

1. **Check you're on the right screen**: The Control tab has 3 screens. Messages appear on the main actions screen.
2. **Messages auto-clear after 3 seconds**: They might have disappeared already if you looked away
3. **Check the terminal supports colors**: The messages use ANSI color codes
4. **Try in mock mode first**: `./bin/imperm-ui --mock` to test without needing a server
