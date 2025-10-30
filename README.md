
<p align="center">
 <img alt="Logo" src = "https://files.catbox.moe/7as151.png"/>
</p>


**A beautiful terminal interface for Letterboxd**
<p align="left">
  <img alt="Main Menu Demo" src="https://files.catbox.moe/up13tf.gif" />
</p>

|Feature|Description|
|---|---|
|**Modern Dashboard**|A beautiful, multi-column main menu with a random movie quote, color palette, and quick-tip sections.|
|**Movie Search**|Search for any movie on Letterboxd and view a detailed, tabbed breakdown of its info, stats, reviews, similar movies, and where to watch.|
|**User Profile**|View any user's profile with tabs for their stats, favorites, recent activity, paginated reviews, and paginated social graph.|
|**List Search**|Find any public list on Letterboxd and browse its contents in a table.|
|**Watchlist Viewer**|View the complete watchlist for any user in a scrollable table.|
|**Diary Viewer**|Browse any user's complete film diary with pagination.|
|**CSV Export**|Export any list, watchlist, or diary to a `.csv` file at a custom, user-specified path.|
|**Help Screen**|A built-in help menu (`?`) for all application keybindings.|
|**Cross-Platform**|Packaged to run on both Linux (Snap, archive) and Windows (archive) with no external dependencies.|

## 🚀 Installation

You can install `LetterCLI` in a few ways. The easiest method for Linux users is via Snap.

### 1. Snap Package (Recommended for Linux)

This is the simplest way to get `LetterCLI` on most Linux distributions (like Ubuntu, Pop!_OS, Debian, Fedora, etc.).

1. Open your terminal and run:
    
    ```
    sudo snap install lettercli
    ```
    
2. Run the app:
    
    ```
    lettercli
    ```
    

_(Note: You will need to have successfully registered and published the 'lettercli' name to the Snap Store for this to work publicly. For your local test build, you would use `sudo snap install --dangerous dist/lettercli_*.snap`)_

### 2. Manual Release (Linux & Windows)

You can download the latest `.tar.gz` (for Linux) or `.zip` (for Windows) file directly from the [GitHub Releases](https://www.google.com/search?q=https://github.com/anshonweb/letterbox-cli/releases "null") page.

1. Download the correct archive for your operating system.
    
2. Extract the archive. You will get a folder with this structure:
    
    ```
    lettercli_v1.0.0_linux_amd64/
    ├── lettercli      # <-- The main application
    ├── py_execs/      # <-- Folder with Python backends
    └── README.md
    ```
    
3. Navigate into the folder with your terminal.
    

**On Linux:**

```
# You may need to make the binaries executable first
chmod +x lettercli
chmod +x py_execs/*

# Run the app
./lettercli
```

**On Windows:**

```
# Open Command Prompt or PowerShell in the extracted folder
.\lettercli.exe
```

**IMPORTANT:** The `py_execs` folder **must** be kept in the same directory as the `lettercli` (or `lettercli.exe`) executable for the application to function.

### 3. From Source (Developers)

If you want to build the project yourself:

1. Clone the repository:
    
    ```
    git clone [https://github.com/anshonweb/letterbox-cli.git](https://github.com/anshonweb/letterbox-cli.git)
    cd letterbox-cli
    ```
    
2. Install dependencies:
    
    - [Go](https://go.dev/doc/install "null") (version 1.18+ recommended)
        
    - [Python 3](https://www.python.org/downloads/ "null")
        
    - [PyInstaller](https://pyinstaller.org/en/stable/installation.html "null") (`pip install pyinstaller`)
        
    - Python libraries: `pip install -r python/scripts/requirements.txt`
        
3. Build the Python executables:
    
    - You must run PyInstaller for all scripts in `python/scripts/`.
        
    - Place the final executables in the correct `dist_py/` folder (e.g., `dist_py/linux_amd64/` for Linux).
        
4. Build and run the Go application:
    
    ```
    go run ./cmd/letterbox-cli
    ```
    

## 📄 License

This project is licensed under the **MIT License**.

See the LICENSE file in the repository for the full text.

## 🙏 Acknowledgements

- [**Charm**](https://github.com/charmbracelet "null") for their incredible Go libraries (Bubble Tea, Lipgloss, Bubbles) that make building beautiful TUIs possible.
    
- [**superstarryeyes/bit**](https://github.com/superstarryeyes/bit "null") for the colored ASCII art logo.