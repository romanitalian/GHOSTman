# GHOSTman

A GUI application for executing HTTP commands, built with Go and Fyne framework. This tool provides a user-friendly interface for managing and executing HTTP requests defined in Postman Collection format.

## Features
- Modern GUI interface built with Fyne
- Command management through Postman Collection files
- Support for multiple command groups
- Command filtering and search
- HTTP request execution with customizable headers and methods
- Response visualization
- Dark/Light theme support (switcher in the top panel)
- Cross-platform (Windows, macOS, Linux)

## Screenshots

- Theme: `Ligth`

![GHOSTman Screenshot](./docs/screenshot-GHOSTman-v1.0.0-theme-light.png)


- Theme: `Dark`

![GHOSTman Screenshot](./docs/screenshot-GHOSTman-v1.0.0-theme-dark.png)

## Prerequisites
- Go 1.21 or higher
- Fyne dependencies:
  - For macOS: Xcode Command Line Tools
  - For Linux: gcc, libgl1-mesa-dev, xorg-dev
  - For Windows: gcc (MinGW-w64)

## Installation

### From Source
```bash
# Clone the repository
git clone https://github.com/romanitalian/GHOSTman.git
cd GHOSTman

# Build the application
go build -o ghostman

# Run the application
./ghostman
```

### Using Go Install
```bash
go install github.com/romanitalian/GHOSTman@latest
```

## Usage

### Configuration
Create a Postman Collection file (e.g., `collection.json`) in the `data` directory:

```json
{
  "info": {
    "name": "My API Collection",
    "description": "Collection of API endpoints"
  },
  "item": [
    {
      "name": "Get Users",
      "request": {
        "method": "GET",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "url": {
          "raw": "https://api.example.com/users",
          "host": ["api.example.com"],
          "path": ["users"]
        }
      }
    }
  ]
}
```

### Running Commands
1. Launch the application
2. Select a command from the left panel
3. (Optional) Switch between Light and Dark theme using the selector at the top
4. Click "Execute" or press Enter
5. View the response in the right panel

## Development

### Setup Development Environment
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
make test-coverage  # or make tc for short

# Build for development
go build -tags dev
```

### Project Structure
```
.
├── data/          # Postman Collection files
├── coverage/      # Test coverage reports
├── main.go        # Application entry point
├── Makefile       # Build and development commands
├── go.mod         # Go module definition
└── go.sum         # Go module checksums
```

## Future Plans

- Implement synchronization via Git
- Add support for gRPC
- AI-powered request suggestions
- Visual API flow builder
- Built-in API documentation generator
- End-to-end API test recorder
- Cloud sync and team collaboration

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: entity - add amazing feature'`)
4. Push to the branch (`git push --force-with-lease origin feature/amazing-feature`)
5. Open a Pull Request
