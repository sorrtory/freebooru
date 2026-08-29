- vue.js frontend + wails.io for a desktop app
- golang core
  - tagging core
  - config parser
  - database cli
- golang http
- golang rclone usage

adapters

- daemon
- cli
- wails

```
Linux .deb:
  /usr/bin/myapp
  /usr/bin/myappd
  /usr/bin/myapp-gui

Windows:
  MyApp.exe
  myapp.exe
  myappd.exe
```

```
myapp/
  cmd/
    myapp/          # CLI entrypoint
    myappd/         # HTTP daemon entrypoint
    myapp-desktop/  # Wails entrypoint, if separate

  internal/
    app/            # application/service layer
    jobs/           # job queue, progress, cancellation
    rclone/         # rclone integration
    tagging/        # file tagging logic
    config/         # config parser/loader
    fsops/          # filesystem operations
    api/            # HTTP handlers
```
