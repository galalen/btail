# btail 🐝 (beautiful) - Interactive File Tail Viewer

btail is a command-line utility for viewing the tail of files with an interactive terminal user interface. It allows you to monitor log files in real-time with features like live updates, search functionality, syntax highlighting, and easy navigation.

<img src="./assets/demo.gif" width="600" alt="btail in action" style="width: 100%;">

## Binary Files
Download binary files from [releases](https://github.com/galalen/btail/releases)

## Installation
Make sure you have [go](https://go.dev/) installed on your system
```bash 
go install github.com/galalen/btail@latest
```

## Usage
Basic usage:
Options:
- `-n <number>`: Set the number of lines to display (default: 5)
- `-f`: Enable follow mode to watch for new lines

Examples:
```bash
btail -n 10 -f path/to/file.log
```

## Special Thanks

- [nxadm/tail](https://github.com/nxadm/tail) and [grafana/tail](https://github.com/grafana/tail): For inspiring the core file tailing functionality.
- [bubbletea](https://github.com/charmbracelet/bubbletea): For powering the interactive terminal UI.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License]
