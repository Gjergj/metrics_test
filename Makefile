# Default target
.PHONY: all
all: download-deps

# Download dependencies target
.PHONY: download-deps
download-deps:
	@echo "Downloading OpenTelemetry Contrib..."
	@mkdir -p deps
	@if [ "$(shell uname)" = "Darwin" ]; then \
		echo "Detected macOS"; \
		curl -L -o otelcontrib.tar.gz "https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/latest/download/otelcontribcol_darwin_amd64.tar.gz" || exit 1; \
		tar -xzf otelcontrib.tar.gz -C deps || exit 1; \
		rm -f otelcontrib.tar.gz || exit 1; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		echo "Detected Linux"; \
		curl -L -o otelcontrib.tar.gz "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.131.0/otelcol-contrib_0.131.0_linux_amd64.tar.gz" || exit 1; \
		tar -xzf otelcontrib.tar.gz -C deps || exit 1; \
		rm -f otelcontrib.tar.gz || exit 1; \
	else \
		echo "Unsupported platform"; \
		exit 1; \
	fi

# Clean target to remove downloaded files
.PHONY: clean
clean:
	rm -rf deps

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all       - Download dependencies (default)"
	@echo "  download-deps - Download OpenTelemetry Contrib"
	@echo "  clean     - Remove downloaded files"
	@echo "  help      - Show this help message"
