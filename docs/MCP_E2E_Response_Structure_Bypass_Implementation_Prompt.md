# MCP E2E Test Content Marshaling Fix - Response Structure Bypass Implementation

## MISSION STATEMENT

Implement a comprehensive solution to fix the MCP content marshaling issue in E2E tests by bypassing direct `mcp.CallToolResult` unmarshaling and working with raw JSON responses. This approach avoids the fundamental JSON interface unmarshaling limitation while maintaining compatibility with the existing MCP v0.29.0 API.

## PROBLEM ANALYSIS

### Current Failure State
- **Error Pattern**: `"failed to decode tool result: json: cannot unmarshal object into Go struct field CallToolResult.content of type mcp.Content"`
- **Scope**: All 22 E2E tests failing with identical marshaling errors
- **Root Cause**: JSON cannot unmarshal into `mcp.Content` interface type without concrete type information
- **Location**: `internal/server/e2e_test.go:109` in `invokeTool` method

### Technical Context
- **Library**: `github.com/mark3labs/mcp-go v0.29.0`
- **Response Creation**: Tools use `mcp.NewToolResultText(string(jsonBytes))` correctly
- **Problem Location**: E2E test framework's response consumption, not response creation
- **Interface Challenge**: `mcp.CallToolResult.Content` is `[]mcp.Content` where `Content` is an interface

### Current Implementation Pattern
```go
// FAILING: Direct unmarshaling into CallToolResult
var toolResult mcp.CallToolResult
if err := json.NewDecoder(resp.Body).Decode(&toolResult); err != nil {
    return nil, fmt.Errorf("failed to decode tool result: %w", err)
}
```

## IMPLEMENTATION STRATEGY: RESPONSE STRUCTURE BYPASS

### Core Approach
Replace direct `mcp.CallToolResult` unmarshaling with raw JSON processing that extracts the actual tool response content without dealing with MCP interface types.

### Implementation Phases

#### Phase 1: Raw JSON Response Processing
- Replace `mcp.CallToolResult` unmarshaling with `map[string]interface{}`
- Extract content from the raw JSON structure
- Create validation functions for response structure

#### Phase 2: Content Extraction Logic
- Implement content extraction from MCP response structure
- Handle both success and error response patterns
- Maintain compatibility with existing test expectations

#### Phase 3: Response Validation Framework
- Create response validation helpers
- Implement content verification methods
- Ensure proper error handling and reporting

#### Phase 4: Test Compatibility Layer
- Update all E2E test expectations
- Maintain existing test logic with new response handling
- Preserve test semantics and validation patterns

## DETAILED IMPLEMENTATION REQUIREMENTS

### 1. Core Response Processing (`internal/server/e2e_test.go`)

#### Replace invokeTool Method
```go
// CURRENT (FAILING):
func (s *E2ETestSuite) invokeTool(toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
    // ... HTTP request logic ...
    var toolResult mcp.CallToolResult
    if err := json.NewDecoder(resp.Body).Decode(&toolResult); err != nil {
        return nil, fmt.Errorf("failed to decode tool result: %w", err)
    }
    return &toolResult, nil
}

// TARGET (NEW IMPLEMENTATION):
func (s *E2ETestSuite) invokeTool(toolName string, params map[string]interface{}) (*ToolResponse, error) {
    // ... HTTP request logic ...
    var rawResponse map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
        return nil, fmt.Errorf("failed to decode raw response: %w", err)
    }
    return extractToolResponse(rawResponse)
}
```

#### Create New Response Types
```go
// New response structure that avoids MCP interface issues
type ToolResponse struct {
    Content    string                 `json:"content"`
    Success    bool                   `json:"success"`
    Message    string                 `json:"message"`
    Metadata   map[string]interface{} `json:"metadata"`
    RawContent map[string]interface{} `json:"raw_content"`
}

// Response extraction function
func extractToolResponse(rawResponse map[string]interface{}) (*ToolResponse, error) {
    // Implementation to extract content from MCP response structure
}
```

### 2. Content Extraction Logic

#### MCP Response Structure Analysis
```go
// Expected MCP response structure:
{
    "content": [
        {
            "type": "text",
            "text": "{\"success\":true,\"message\":\"...\",\"formattedCode\":\"...\"}"
        }
    ]
}

// Extraction logic:
func extractToolResponse(rawResponse map[string]interface{}) (*ToolResponse, error) {
    // 1. Navigate to content array
    // 2. Extract text content from first element
    // 3. Parse inner JSON containing actual tool response
    // 4. Create ToolResponse with extracted data
}
```

#### Content Processing Pipeline
1. **Raw Response Validation**: Verify basic MCP structure
2. **Content Array Extraction**: Navigate to `content` field
3. **Text Content Extraction**: Extract `text` field from content elements
4. **Inner JSON Parsing**: Parse the actual tool response JSON
5. **Response Object Creation**: Create `ToolResponse` with extracted data

### 3. Response Validation Framework

#### Validation Functions
```go
// Validate MCP response structure
func validateMCPStructure(rawResponse map[string]interface{}) error {
    // Verify required fields exist
    // Check content array structure
    // Validate text content format
}

// Extract and validate content
func extractAndValidateContent(rawResponse map[string]interface{}) (string, error) {
    // Navigate MCP structure safely
    // Extract text content with error handling
    // Validate content is valid JSON
}

// Parse tool response content
func parseToolContent(contentText string) (map[string]interface{}, error) {
    // Parse inner JSON safely
    // Validate tool response structure
    // Return parsed content
}
```

### 4. Test Compatibility Updates

#### Update Test Expectations
- Modify all test assertions to use new `ToolResponse` structure
- Update content validation logic
- Maintain existing test semantics

#### Preserve Test Logic
```go
// BEFORE:
result, err := s.invokeTool("go_fmt", params)
require.NoError(t, err)
// Access result.Content[0].Text and parse JSON

// AFTER:
result, err := s.invokeTool("go_fmt", params)
require.NoError(t, err)
// Direct access to result.Content, result.Success, etc.
```

## IMPLEMENTATION FILES

### Primary Files to Modify
1. **`internal/server/e2e_test.go`**
   - Replace `invokeTool` method
   - Add new response types
   - Add content extraction functions
   - Update all test methods

2. **`internal/server/comprehensive_e2e_test.go`**
   - Update test methods to use new response structure
   - Modify assertions for new response format
   - Preserve test validation logic

### Supporting Files (If Needed)
3. **`internal/testing/helpers.go`**
   - Add response validation utilities
   - Create content extraction helpers
   - Add debugging/logging functions

4. **New File: `internal/testing/response_utils.go`**
   - Centralize response processing logic
   - Provide reusable extraction functions
   - Handle edge cases and error scenarios

## TECHNICAL REQUIREMENTS

### 1. Error Handling
- Comprehensive error handling for malformed responses
- Clear error messages for debugging
- Graceful handling of unexpected response structures

### 2. Backward Compatibility
- Maintain existing test semantics
- Preserve all test validation logic
- Ensure no loss of test coverage

### 3. Performance Considerations
- Minimize JSON parsing overhead
- Efficient content extraction
- Avoid unnecessary object creation

### 4. Code Quality Standards
- Follow Go idioms and best practices
- Proper error handling patterns
- Comprehensive documentation
- Unit tests for new extraction logic

## VALIDATION CRITERIA

### 1. E2E Test Success
- All 22 E2E tests must pass
- No regression in test functionality
- Proper assertion of tool responses

### 2. Response Content Validation
- Successful extraction of tool response content
- Proper handling of both success and error responses
- Accurate content parsing and validation

### 3. Error Handling Verification
- Proper handling of malformed responses
- Clear error reporting for debugging
- Graceful degradation for edge cases

### 4. Performance Validation
- No significant performance degradation
- Efficient JSON processing
- Minimal memory overhead

## SUCCESS METRICS

### Immediate Success Indicators
- [ ] All E2E tests pass without marshaling errors
- [ ] Content extraction works for all tool types
- [ ] Error responses handled correctly
- [ ] No regression in existing test logic

### Long-term Success Indicators
- [ ] Maintainable and extensible response handling
- [ ] Clear separation between MCP structure and tool content
- [ ] Robust error handling for edge cases
- [ ] Comprehensive test coverage for new functionality

## IMPLEMENTATION APPROACH

### Parallel Development Threads

#### Thread 1: Core Response Processing
- Implement raw JSON unmarshaling
- Create content extraction pipeline
- Add response validation logic

#### Thread 2: Response Type Design
- Design new `ToolResponse` structure
- Implement content parsing methods
- Create validation helpers

#### Thread 3: Test Compatibility
- Update test assertions
- Modify test expectations
- Preserve test semantics

#### Thread 4: Error Handling
- Implement comprehensive error handling
- Add debugging capabilities
- Create recovery mechanisms

### Implementation Order
1. **Foundation**: Create new response types and extraction functions
2. **Core Logic**: Implement `invokeTool` replacement with raw JSON processing
3. **Content Processing**: Add content extraction and validation pipeline
4. **Test Updates**: Update all E2E tests to use new response structure
5. **Validation**: Comprehensive testing and validation of new approach
6. **Documentation**: Update code documentation and comments

## DEBUGGING AND TESTING STRATEGY

### Debugging Capabilities
- Raw response logging for troubleshooting
- Content extraction tracing
- Validation step debugging
- Error context preservation

### Testing Approach
- Unit tests for content extraction functions
- Integration tests for response processing
- E2E test validation for all tool types
- Error handling test coverage

## CONSTRAINTS AND CONSIDERATIONS

### Must Preserve
- All existing test functionality
- Tool response validation logic
- Error handling semantics
- Test assertion patterns

### Must Avoid
- Breaking changes to test interfaces
- Performance degradation
- Loss of test coverage
- Complexity in test logic

### Future Considerations
- Extensibility for new tool types
- Maintainability of response processing
- Compatibility with MCP library updates
- Migration path for future improvements

## COMPLETION CHECKLIST

### Phase 1: Foundation
- [ ] New response types defined
- [ ] Content extraction functions implemented
- [ ] Basic validation logic created

### Phase 2: Core Implementation
- [ ] `invokeTool` method replaced
- [ ] Raw JSON processing implemented
- [ ] Content extraction pipeline working

### Phase 3: Test Updates
- [ ] All E2E tests updated
- [ ] Test assertions modified
- [ ] Test validation preserved

### Phase 4: Validation
- [ ] All 22 E2E tests passing
- [ ] Error handling verified
- [ ] Performance validated
- [ ] Documentation updated

## NOTES

This implementation avoids the fundamental JSON interface unmarshaling issue by working with the raw response structure directly. While it sacrifices some type safety, it provides a robust solution that bypasses the MCP `Content` interface limitation entirely. The approach maintains full compatibility with the existing MCP v0.29.0 API and preserves all test functionality while solving the marshaling problem comprehensively.
