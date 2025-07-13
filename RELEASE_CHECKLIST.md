# Axolotl Game - Release Checklist

Use this checklist when preparing a release of your game.

## 📋 Pre-Release Testing

- [ ] **Code Compilation**: All platforms build without errors
- [ ] **Game Mechanics**: All features work correctly
  - [ ] Player movement (WASD/Arrow keys)
  - [ ] Attack system (Q key)
  - [ ] Health regeneration
  - [ ] Slime spawning
  - [ ] Kill counter
  - [ ] Death/reset system
- [ ] **Assets**: All sprites and textures load properly
- [ ] **Performance**: Game runs at 60 FPS on target systems

## 🔨 Build Process

- [ ] **Clean Build Environment**: `go mod tidy` and `go mod download`
- [ ] **Cross-Platform Builds**:
  - [ ] Windows 64-bit (`axolotl-windows.exe`)
  - [ ] macOS Intel (`axolotl-macos-intel`)
  - [ ] macOS Apple Silicon (`axolotl-macos-arm64`)
  - [ ] Linux 64-bit (`axolotl-linux`)
- [ ] **Asset Copying**: Each platform has its own assets folder
- [ ] **File Permissions**: macOS/Linux executables are executable

## 📦 Distribution Packages

Create separate packages for each platform:

### Windows Package
- [ ] `axolotl-windows-v[VERSION].zip`
  - [ ] `axolotl-windows.exe`
  - [ ] `assets/` folder (complete)
  - [ ] `README.txt` (player instructions)

### macOS Intel Package  
- [ ] `axolotl-macos-intel-v[VERSION].zip`
  - [ ] `axolotl-macos-intel` (executable)
  - [ ] `assets/` folder (complete)
  - [ ] `README.txt` (player instructions)

### macOS Apple Silicon Package
- [ ] `axolotl-macos-arm64-v[VERSION].zip`
  - [ ] `axolotl-macos-arm64` (executable) 
  - [ ] `assets/` folder (complete)
  - [ ] `README.txt` (player instructions)

### Linux Package
- [ ] `axolotl-linux-v[VERSION].zip`
  - [ ] `axolotl-linux` (executable)
  - [ ] `assets/` folder (complete)
  - [ ] `README.txt` (player instructions)

## 📄 Documentation

- [ ] **Player README**: Simple instructions for players
- [ ] **Controls**: Movement, attack, special keys
- [ ] **Objective**: How to play and win
- [ ] **System Requirements**: Minimum specs for each platform
- [ ] **Troubleshooting**: Common issues and solutions

## 🧪 Quality Assurance

- [ ] **Test Each Package**: Download and test each platform package
- [ ] **Fresh System Testing**: Test on clean systems if possible
- [ ] **Performance Check**: Verify smooth gameplay
- [ ] **Asset Verification**: All sprites appear correctly
- [ ] **Save/Load**: No save system, but verify game resets work

## 🚀 Release Preparation

- [ ] **Version Number**: Update version in all documentation
- [ ] **Release Notes**: Document new features/fixes
- [ ] **Screenshots**: Current gameplay screenshots
- [ ] **File Checksums**: Generate SHA256 hashes for security
- [ ] **Virus Scan**: Scan all packages with antivirus

## 📢 Distribution Channels

Choose your distribution method:

- [ ] **GitHub Releases**: Create tagged release with binaries
- [ ] **Itch.io**: Upload to itch.io for easy distribution
- [ ] **Personal Website**: Host files on your own site
- [ ] **Game Forums**: Share in indie game communities

## ✅ Final Checks

- [ ] **All Packages Created**: Every platform has a complete package
- [ ] **Documentation Complete**: All README files included
- [ ] **Assets Verified**: All assets present and working
- [ ] **Testing Complete**: Each package tested on target platform
- [ ] **Version Consistent**: Same version number everywhere
- [ ] **Backup Created**: Source code and assets backed up

## 🎯 Post-Release

- [ ] **Monitor Feedback**: Watch for bug reports
- [ ] **Update Documentation**: Fix any unclear instructions  
- [ ] **Plan Next Version**: Based on player feedback
- [ ] **Marketing**: Share on social media, forums, etc.

---

**Release Version**: v___________  
**Release Date**: ___________  
**Released By**: ___________

## Quick Build Commands

```bash
# Build all platforms
./build.sh

# Or with make
make all-platforms

# Test current platform
go run cmd/main.go
``` 