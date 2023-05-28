# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOINSTALL = $(GOCMD) install

# Directories
BUILD_DIR = build/

# Build targets
TARGETS := $(wildcard cmd/*/main.go)
BINARY_NAMES := $(patsubst cmd/%/main.go,%,$(TARGETS))
BINARY_PATHS := $(addprefix $(BUILD_DIR)/,$(BINARY_NAMES))

# Default target
all: build

# Create the build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build the targets
build: $(BUILD_DIR) $(BINARY_PATHS)

$(BUILD_DIR)/%: cmd/%/main.go
	$(GOBUILD) -o $@ $<

# Clean the binaries
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Install the targets
install: build
	$(GOINSTALL) ./...

# Phony targets
.PHONY: all build clean install
