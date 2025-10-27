package utils

// GetTradeAnalysisPrompt returns the prompt template for trade analysis
func GetTradeAnalysisPrompt() string {
	return `Analyze the following trade data based on the user's prompt.
If the user's prompt requires specific information or analysis directly from the trade data, integrate that information prominently into your response.
Maintain a helpful, informative, and professional tone suitable for financial analysis.
--- Rules ---
- You MUST use Markdown format for your response.
- You MUST use the trade data to answer the user's request.
--- Trade Data Provided ---
%s
--- End of Trade Data ---

--- User's Request ---
%s`
}
