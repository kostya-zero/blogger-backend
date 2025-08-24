set shell := ["bash", "-c"]
set windows-shell := ["pwsh.exe", "-NoLogo", "-Command"]

binaryPath := if os() == "windows" { './build/blogger.exe' } else { './build/blogger' }

# Runs build recipe
default: build

# Update dependencies
update:
    go get -u ./...

# Build the project to an executable
build:
    go build -o {{ binaryPath }} cmd/blogger/main.go

# Remove the build artifacts
clean:
    rm -f ./build

# Run the application with optional arguments
run *ARGS:
    go run cmd/blogger/main.go {{ ARGS }}
