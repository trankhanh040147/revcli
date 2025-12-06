package image

import "github.com/trankhanh040147/revcli/internal/gemini"

// EstimateCost estimates the cost for an image generation request
// based on prompt text length and resolution, before actual generation
func EstimateCost(prompt, resolution string) *gemini.CostBreakdown {
	// Estimate prompt tokens: ~1 token per 4 characters (conservative estimate)
	// Minimum 1 token for any non-empty prompt
	promptLength := len(prompt)
	estimatedPromptTokens := int32(promptLength / 4)
	if estimatedPromptTokens < 1 && promptLength > 0 {
		estimatedPromptTokens = 1
	}

	// Calculate input cost: (estimatedPromptTokens / 1,000,000) * $2.00
	inputCost := (float64(estimatedPromptTokens) / 1_000_000.0) * InputPricePerMillionTokens

	// Output cost is fixed based on resolution (image output)
	var outputCost float64
	switch resolution {
	case Resolution4K:
		outputCost = ImagePrice4K
	case Resolution1K, Resolution2K:
		outputCost = ImagePrice1K
	default:
		// Fallback: use 1K pricing for unknown resolutions
		outputCost = ImagePrice1K
	}

	totalCost := inputCost + outputCost

	// Calculate VND costs
	inputCostVND := inputCost * USDToVNDExchangeRate
	outputCostVND := outputCost * USDToVNDExchangeRate
	totalCostVND := totalCost * USDToVNDExchangeRate

	return &gemini.CostBreakdown{
		InputCost:     inputCost,
		OutputCost:    outputCost,
		TotalCost:     totalCost,
		InputCostVND:  inputCostVND,
		OutputCostVND: outputCostVND,
		TotalCostVND:  totalCostVND,
	}
}

// CalculateCost calculates the cost for an image generation request
// based on token usage and resolution
func CalculateCost(promptTokens, completionTokens int32, resolution string, hasImageOutput, hasTextOutput bool) *gemini.CostBreakdown {
	if promptTokens < 0 || completionTokens < 0 {
		return &gemini.CostBreakdown{}
	}

	// Calculate input cost: (promptTokens / 1,000,000) * $2.00
	inputCost := (float64(promptTokens) / 1_000_000.0) * InputPricePerMillionTokens

	var outputCost float64

	// Calculate output cost based on what was generated
	if hasImageOutput {
		// For images, use fixed pricing based on resolution
		switch resolution {
		case Resolution4K:
			outputCost = ImagePrice4K
		case Resolution1K, Resolution2K:
			outputCost = ImagePrice1K
		default:
			// Fallback: calculate based on token count if resolution unknown
			outputCost = (float64(completionTokens) / 1_000_000.0) * OutputImagePricePerMillionTokens
		}
	} else if hasTextOutput {
		// For text output: (completionTokens / 1,000,000) * $12.00
		outputCost = (float64(completionTokens) / 1_000_000.0) * OutputTextPricePerMillionTokens
	}

	// If both image and text, we need to handle separately
	// The API typically returns image tokens in completionTokens, but we use fixed pricing for images
	// If there's text output, we'd need to know how many tokens were text vs image
	// For now, we prioritize image pricing if image exists, otherwise use text pricing

	totalCost := inputCost + outputCost

	// Calculate VND costs
	inputCostVND := inputCost * USDToVNDExchangeRate
	outputCostVND := outputCost * USDToVNDExchangeRate
	totalCostVND := totalCost * USDToVNDExchangeRate

	return &gemini.CostBreakdown{
		InputCost:     inputCost,
		OutputCost:    outputCost,
		TotalCost:     totalCost,
		InputCostVND:  inputCostVND,
		OutputCostVND: outputCostVND,
		TotalCostVND:  totalCostVND,
	}
}
