<p align="center">
  <img src="./logo.png" alt="GitCAD" width="180" />
</p>

# GitCAD

**Version control system designed for CAD files** — with git-like CLI and a desktop GUI.

Supports DWG, STL, DXF, and OBJ files with intelligent geometry-aware diffing. Shows what changed in your 3D models and architectural floor plans with red (removed) and green (added) highlighting.

![GitCAD Desktop GUI showing DWG diffing](./screenshot.png)

## Quick Start

### Build the CLI

```bash
cd /path/to/gitcad
go build -o gitcad .
```

This produces a single `gitcad` binary. Add it to your PATH:

```bash
export PATH=$PWD:$PATH
```

### Basic Workflow

```bash
# Initialize a new repository
mkdir my-cad-project && cd my-cad-project
gitcad init

# Add and commit files
gitcad add model.stl drawing.dxf
gitcad commit -m "Initial design"

# Create a branch and make changes
gitcad branch feature-chamfer
gitcad checkout feature-chamfer

# Edit your CAD files, then see what changed
gitcad diff          # Red/green colored output

# Commit and merge back
gitcad add .
gitcad commit -m "Added chamfer to edge"
gitcad checkout main
gitcad merge feature-chamfer
```

### All Commands

| Command | Description |
|---------|-------------|
| `gitcad init` | Initialize repository (default: `main` branch) |
| `gitcad add <files>` | Stage files for commit |
| `gitcad commit -m "msg"` | Commit staged changes |
| `gitcad status` | Show working tree status |
| `gitcad log [-n N]` | Show commit history |
| `gitcad branch [name]` | List or create branches |
| `gitcad branch -d name` | Delete a branch |
| `gitcad checkout <branch>` | Switch branches |
| `gitcad diff [file]` | Show changes with red/green coloring |
| `gitcad merge <branch>` | Merge branch into current |

## Desktop GUI (Tauri)

### Prerequisites

- [Rust](https://rustup.rs/) (≥ 1.77)
- [Node.js](https://nodejs.org/) (≥ 18)

### Run in Development

```bash
cd gui
npm install
npx @tauri-apps/cli dev
```

### Build Distributable App

```bash
cd gui
npm install
npx @tauri-apps/cli build
```

This produces:
- **macOS**: `.dmg` installer
- **Windows**: `.msi` installer
- **Linux**: `.deb` / `.AppImage`

## CAD-Aware Diffing

GitCAD understands CAD file formats:

| Format | What's Compared |
|--------|----------------|
| **STL** | Triangle count, bounding box, surface area, per-triangle changes |
| **DXF** | Entity types (LINE, ARC, CIRCLE...), layers, entity counts |
| **DWG** | Converted to DXF via LibreDWG; Entity types, layers, entity counts |
| **OBJ** | Vertex count, face count, normals, groups, materials |
| **Text** | Line-by-line unified diff |
| **Binary** | File size changes |

## Architecture

```
gitcad/
├── main.go          # CLI entry point
├── cmd/             # Cobra CLI commands
├── core/            # VCS engine (objects, staging, refs, merge)
├── diff/            # CAD-aware diff engine (STL, DXF, OBJ)
└── gui/             # Tauri desktop app (React + Three.js)
```

### Storage Model

Like Git, GitCAD uses content-addressable storage:

- **Blobs**: File content compressed with zlib, addressed by SHA-256
- **Trees**: Directory listings mapping names to blob/tree hashes
- **Commits**: Tree hash + parent(s) + author + timestamp + message
- **Objects stored at**: `.gitcad/objects/<hash[:2]>/<hash[2:]>`

## License

MIT
