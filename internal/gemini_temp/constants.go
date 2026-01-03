package gemini_temp

import "google.golang.org/genai"

// DefaultHistoryCapacity is the default capacity for pre-allocating chat history slice
const DefaultHistoryCapacity = 20

// DefaultModelTemperature is the default temperature parameter
const DefaultModelTemperature = float32(0.3)

// DefaultModelTopP is the default topP parameter
const DefaultModelTopP = float32(0.95)

// DefaultModelTopK is the default topK parameter
const DefaultModelTopK = 40

// StreamChannelBufferSize is the buffer size for streaming result channels
// Buffer of 20 allows producer (goroutine) to continue without blocking
// while consumer processes results, improving streaming performance
const StreamChannelBufferSize = 20

// ChunkChannelBufferSize is the buffer size for chunk channels in UI streaming
// Buffer of 10 provides smooth chunk delivery without blocking the producer
const ChunkChannelBufferSize = 10

// ErrorChannelBufferSize is the buffer size for error channels
// Buffer of 1 is sufficient as only one error is sent per operation
const ErrorChannelBufferSize = 1

// DoneChannelBufferSize is the buffer size for completion channels
// Buffer of 1 is sufficient as only one completion signal is sent per operation
const DoneChannelBufferSize = 1

// mapSafetyThreshold maps a string threshold to genai.HarmBlockThreshold
// Supported values: "HIGH", "MEDIUM_AND_ABOVE", "LOW_AND_ABOVE", "NONE", "OFF"
// Defaults to HarmBlockThresholdBlockOnlyHigh if unknown
func mapSafetyThreshold(threshold string) genai.HarmBlockThreshold {
	switch threshold {
	case "HIGH":
		return genai.HarmBlockThresholdBlockOnlyHigh
	case "MEDIUM_AND_ABOVE":
		return genai.HarmBlockThresholdBlockMediumAndAbove
	case "LOW_AND_ABOVE":
		return genai.HarmBlockThresholdBlockLowAndAbove
	case "NONE":
		return genai.HarmBlockThresholdBlockNone
	case "OFF":
		return genai.HarmBlockThresholdOff
	default:
		return genai.HarmBlockThresholdBlockOnlyHigh
	}
}

// getAllHarmCategories returns all harm categories that should have safety settings applied
func getAllHarmCategories() []genai.HarmCategory {
	return []genai.HarmCategory{
		genai.HarmCategoryHarassment,
		genai.HarmCategoryHateSpeech,
		genai.HarmCategorySexuallyExplicit,
		genai.HarmCategoryDangerousContent,
		genai.HarmCategoryCivicIntegrity,
	}
}
