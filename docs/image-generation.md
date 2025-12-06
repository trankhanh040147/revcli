# Image Generation Feature Development

## Overview

The image generation feature allows users to generate images from text prompts using Google's Gemini 3 Pro Image Preview model (`gemini-3-pro-image-preview`). This feature is implemented as a separate command to keep it isolated from the main code review functionality.

## Current Status

**Status:** ✅ Implemented (v0.4)

**Features Completed:**
- [x] CLI command `generate-image` (aliases: `image`, `img`)
- [x] Support for text prompts
- [x] Configurable aspect ratios (1:1, 2:3, 3:2, 4:3, 3:4, 16:9, 9:16)
- [x] Configurable resolutions (1K, 2K, 4K)
- [x] Custom output file paths
- [x] Token usage display
- [x] Cost calculation and display (USD)
- [x] Cost display in VND (Vietnamese Dong)
- [x] Cost estimation before generation (USD and VND)
- [x] Image data extraction and saving
- [x] Text description extraction (if provided by API)

## Implementation Details

### Architecture

The feature is organized in dedicated files to maintain separation from code review functionality:

```
internal/
  image/
    constants.go    # Model, aspect ratios, resolutions, pricing constants
    cost.go         # Cost calculation logic
    generator.go    # Image generation wrapper
  gemini/
    client.go       # Extended with GenerateImage() method
cmd/
  generate-image.go # CLI command implementation
```

### Key Components

#### 1. Constants (`internal/image/constants.go`)
- Model name: `gemini-3-pro-image-preview`
- Supported aspect ratios and resolutions
- Pricing constants based on [official Gemini 3 Pro Image Preview pricing](https://ai.google.dev/gemini-api/docs/pricing#gemini-3-pro-image-preview)
- USD to VND exchange rate

#### 2. Cost Calculation (`internal/image/cost.go`)
- `EstimateCost()` function estimates cost before generation:
  - Estimates input tokens from prompt length (~1 token per 4 characters)
  - Uses fixed image output pricing based on resolution
  - Returns estimated `CostBreakdown` in USD and VND
- `CalculateCost()` function computes actual cost after generation:
  - Input tokens: $2.00 per 1M tokens
  - Output text tokens: $12.00 per 1M tokens
  - Output images: Fixed pricing by resolution
    - 1K/2K images: $0.134 per image (1120 tokens)
    - 4K images: $0.24 per image (2000 tokens)
- Returns `CostBreakdown` with input, output, and total costs
- Supports both USD and VND currency display

#### 3. Gemini Client Extension (`internal/gemini/client.go`)
- `GenerateImage()` method uses REST API directly (old genai package doesn't support Gemini 3 features)
- Extracts image data from base64-encoded response
- Extracts text descriptions
- Tracks token usage
- Returns `ImageGenerationResult` with image data, text, usage, and cost

#### 4. Generator Wrapper (`internal/image/generator.go`)
- Validates aspect ratios and resolutions
- Calls Gemini client for image generation
- Calculates cost after generation
- Handles file saving with directory creation
- Returns `GenerateImageResult` with all metadata

#### 5. CLI Command (`cmd/generate-image.go`)
- Command: `revcli generate-image` (aliases: `image`, `img`)
- Flags:
  - `--prompt` / `-p`: Text prompt (required)
  - `--output` / `-o`: Output file path (default: `image.png`)
  - `--aspect-ratio` / `-a`: Aspect ratio (default: `1:1`)
  - `--resolution` / `-r`: Resolution (default: `1K`)
  - `--interactive` / `-i`: Interactive mode (placeholder)
  - `--no-interactive` / `-I`: Force CLI-only mode
- Displays:
  - Estimated cost before generation (USD and VND)
  - Success message with file path
  - Text description (if available)
  - Token usage breakdown
  - Actual cost breakdown in USD and VND (for comparison with estimate)

## Pricing Reference

Based on [Gemini 3 Pro Image Preview pricing](https://ai.google.dev/gemini-api/docs/pricing#gemini-3-pro-image-preview):

### Standard Pricing (per 1M tokens in USD)
- **Input**: $2.00 (text/image)
- **Output Text**: $12.00 (text and thinking)
- **Output Images**: $120.00 per 1M tokens

### Fixed Image Pricing
- **1K/2K images**: $0.134 per image (1120 tokens)
- **4K images**: $0.24 per image (2000 tokens)

### Currency Conversion
- USD to VND exchange rate: Configurable constant (default: ~25,000 VND/USD)
- Cost displayed in both USD and VND for convenience

## Usage Examples

```bash
# Basic usage with default settings
revcli generate-image --prompt "A beautiful sunset over mountains"

# Custom aspect ratio and resolution
revcli generate-image -p "A cat wearing sunglasses" -a 16:9 -r 2K -o cat.png

# High resolution 4K image
revcli generate-image -p "Futuristic cityscape" -r 4K -o city.png
```

## Output Format

```
Estimated cost:
  $0.0013 (Input: $0.0003, Output: $0.0010)
  ~32.5 VND (Input: ~7.5 VND, Output: ~25.0 VND)

Generating image...
Prompt: A beautiful sunset over mountains
Aspect Ratio: 1:1
Resolution: 1K

✓ Image generated successfully!
  Saved to: image.png

Description: [Text description from API if available]

📊 Token Usage: 150 prompt + 1120 completion = 1270 total
💰 Cost: $0.0013 (Input: $0.0003, Output: $0.0010)
   ~32.5 VND (Input: ~7.5 VND, Output: ~25.0 VND)
```

## Technical Decisions

### Why REST API Instead of genai Package?
The deprecated `github.com/google/generative-ai-go/genai` package doesn't support Gemini 3 image generation features (response modalities, image config). We use the REST API directly for this feature.

### Why Separate Files?
The image generation feature is unrelated to code review functionality. Keeping it in dedicated files maintains code organization and makes it easier to maintain or potentially extract as a separate module.

### Cost Calculation Approach
- Uses fixed pricing for images (more accurate than token-based for images)
- Falls back to token-based calculation for unknown resolutions
- Handles both image and text outputs separately

### Cost Estimation Approach
- Estimates input tokens using heuristic: ~1 token per 4 characters
- Uses fixed image output pricing based on resolution (same as actual cost)
- Provides cost preview before generation to help users make informed decisions
- Estimation is conservative (typically slightly overestimates) to avoid surprises

## Future Enhancements

### Planned Features
- [ ] Interactive TUI mode for prompt input and parameter selection
- [ ] Batch image generation from multiple prompts
- [ ] Image preview in terminal (if supported)
- [x] Cost estimation before generation
- [ ] Support for additional image models as they become available
- [ ] Image editing/refinement capabilities

### Potential Improvements
- [ ] Configurable exchange rate (via config file or flag)
- [ ] Historical cost tracking
- [ ] Cost limits/warnings
- [ ] Multiple output formats (JPEG, WebP, etc.)
- [ ] Image metadata extraction and display

## Known Limitations

1. **Interactive Mode**: Not yet implemented (placeholder in code)
2. **Multiple Images**: Currently only handles the first image if API returns multiple
3. **Exchange Rate**: Static constant, should be updated periodically or made configurable
4. **Error Handling**: Could be more detailed for specific API error cases

## Testing

### Manual Testing Checklist
- [x] Basic image generation with default settings
- [x] Custom aspect ratios work correctly
- [x] Different resolutions (1K, 2K, 4K) generate correctly
- [x] Token usage is accurate
- [x] Cost calculation is correct for different resolutions
- [x] VND conversion displays correctly
- [x] Cost estimation displays before generation
- [x] Cost estimation is reasonably accurate
- [x] File saving works with custom paths
- [x] Directory creation for nested paths
- [x] Error handling for invalid parameters
- [x] Error handling for API failures

## Related Files

- `internal/image/constants.go` - Constants and pricing
- `internal/image/cost.go` - Cost calculation
- `internal/image/generator.go` - Generator wrapper
- `internal/gemini/client.go` - Gemini client extension
- `cmd/generate-image.go` - CLI command
- `docs/DEVELOPMENT.md` - Main development roadmap

## References

- [Gemini 3 Developer Guide](https://ai.google.dev/gemini-api/docs/gemini-3)
- [Image Generation Documentation](https://ai.google.dev/gemini-api/docs/image-generation)
- [Gemini 3 Pro Image Preview Pricing](https://ai.google.dev/gemini-api/docs/pricing#gemini-3-pro-image-preview)

