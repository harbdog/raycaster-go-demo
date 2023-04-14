# raycaster-go-demo

Demo project for using [raycaster-go](https://github.com/harbdog/raycaster-go) engine as a module.

To see it in action, see the [demo video on YouTube](https://www.youtube.com/watch?v=WKVsmkmYN24).

![Screenshot](docs/images/screenshot.jpg?raw=true)

## How to try

The demo is now available to try in the browser: https://harbdog.github.io/raycaster-go-demo/

- The browser version may run much slower than running locally as an application,
  and may not run very well on old or slow machines. It also requires a mouse and keyboard,
  see the controls listed below.

## How to run

To run the demo from source locally:

1. Download, install, and setup Golang https://golang.org/dl/
2. Clone/download the demo project locally.
3. From the demo project folder, use the following command to run it:
    * `go run main.go`

**NOTE**: Depending on the OS, the Ebitengine game library may have
[additional dependencies to install](https://ebiten.org/documents/install.html).

## Controls

* Press `Escape` or `F1` key to show demo settings menu (also to exit the game)
* Move the mouse to rotate and pitch view
* Move and strafe using `WASD` or `Arrow Keys`
* Click left mouse button to fire current weapon
* Use mouse wheel or press `1` or `2` to select a weapon
* Press `H` to holster/put away current weapon
* Hold `Shift` key to move faster
* Hold `C` key for crouch position
* Hold `Z` key for prone position
* Hold `Spacebar` for jump position
* Hold `ALT` key to enter mouse move mode (vertical mouse moves position instead of pitch)
* Hold `CTRL` key to release mouse cursor capture
