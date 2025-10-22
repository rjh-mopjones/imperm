# Quickstart Guide

## Running the Application

1. **Build the application:**
   ```bash
   go build -o imperm ./cmd/imperm
   ```

2. **Run in mock mode:**
   ```bash
   ./imperm --mock
   ```

## What You'll See

### Control Tab (Default)
- **Left Panel**: Three actions you can perform
  - Build Environment
  - Build Environment with Options
  - Destroy Environment
- **Right Panel**: History of environments you've launched

### Observe Tab (Press Tab to switch)
- **k9s-style interface** showing:
  - Environments (press `e`)
  - Pods (press `p`)
  - Deployments (press `d`)
  - Auto-refreshing every 5 seconds
  - Table view with columns for Name, Namespace, Status, Age, etc.
- **Drill-down capability**:
  - Press `Enter` on an environment to view its pods and deployments
  - Breadcrumb navigation shows your current context
  - Press `Esc` to go back to all environments

## Try It Out

1. Launch the app: `./imperm --mock`
2. You'll start in the **Control** tab
3. Navigate with `↑/↓` or `k/j` keys
4. Press `Enter` on "Build Environment"
5. Type a name like `my-test-env` and press `Enter`
6. See it appear in the history on the right
7. Press `Tab` to switch to the **Observe** tab
8. See the mock environments and pods in a table
9. Press `e` for Environments view, `p` for Pods view
10. Press `a` to toggle auto-refresh
11. Press `q` to quit

## Mock Data

The mock mode comes with pre-populated data:
- 2 environments (dev-env-1, staging-env-1)
- 3 pods across different namespaces with CPU and Memory metrics
- 2 deployments (app-deployment, nginx-deployment)
- Environment history with timestamps

## Drill-Down Workflow

1. In the Observe tab, start on the Environments view
2. Use `↑/↓` or `k/j` to select an environment
3. Press `Enter` to drill down into that environment
4. The breadcrumb at the top shows: `Environments > [env-name] > Pods`
5. Press `p` to view Pods or `d` to view Deployments for that environment
6. Press `Esc` to go back to all environments
7. You can also press `e`, `p`, or `d` to switch views globally (all resources)

## Next Steps

To connect to your real middleware API:

1. Edit `internal/middleware/http.go`
2. Implement the API calls to your endpoints
3. Update `cmd/imperm/main.go` to use HTTPClient instead of MockClient:
   ```go
   client = middleware.NewHTTPClient("https://your-api.com")
   ```

## Key Files to Customize

| File | Purpose |
|------|---------|
| `internal/middleware/http.go` | Implement your API calls here |
| `internal/models/environment.go` | Adjust data models if needed |
| `internal/ui/observe.go` | Customize the k9s-style view |
| `internal/ui/control.go` | Customize the control panel |

## Tips

- The UI auto-refreshes every 5 seconds by default (configurable in `internal/ui/observe.go`)
- Navigation is vim-style (`j`/`k`) or arrow keys
- The history persists in memory during the session
- All components are modular and easy to extend
