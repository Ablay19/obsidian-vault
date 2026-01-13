# Opencode Agents Configuration

This directory contains configurations and specifications for all AI agents available in Opencode.

## Available Agent Types

### 🧠 **General Agent**
- **Purpose**: Research complex questions and execute multi-step tasks
- **Capabilities**: 
  - Multi-unit parallel execution
  - Complex task breakdown
  - Research and analysis
  - Problem solving
- **Max Tokens**: 10,000
- **Default Timeout**: 5 minutes

### 🔍 **Explore Agent**
- **Purpose**: Fast codebase exploration and pattern finding
- **Capabilities**:
  - Rapid file pattern matching
  - Code search and analysis
  - Architecture understanding
  - Function location
- **Thoroughness Levels**:
  - `quick` - Basic searches for obvious patterns
  - `medium` - Moderate exploration with multiple search strategies  
  - `very thorough` - Comprehensive analysis across multiple locations
- **Default Tokens**: 5,000

## Agent Configuration Structure

Each agent configuration includes:
- `name` - Agent identifier
- `description` - Purpose and capabilities
- `tools` - Available tools and permissions
- `parameters` - Default behavior settings
- `capabilities` - What the agent can do
- `limitations` - What the agent cannot do
- `examples` - Usage scenarios

## Agent Specializations

### **Code Assistant Agent**
- Languages: JavaScript, Python, Go, TypeScript, Java
- Frameworks: React, Node.js, Django, Gin, Express
- Capabilities: Debugging, refactoring, optimization, testing

### **Research Agent**  
- Sources: Documentation, Stack Overflow, GitHub, academic papers
- Capabilities: Technical research, trend analysis, best practices
- Output: Structured reports with sources and citations

### **Creative Writing Agent**
- Styles: Technical documentation, blog posts, tutorials
- Capabilities: Content creation, documentation, explanations
- Formats: Markdown, HTML, plain text

## Agent Workflow

1. **Task Assignment** - Receive task description and context
2. **Planning** - Break down task into actionable steps
3. **Execution** - Use available tools to complete task
4. **Validation** - Review and validate results
5. **Reporting** - Return results with metadata

## Agent Collaboration

Agents can:
- Handoff tasks between specialists
- Share context and findings
- Work in parallel on different aspects
- Validate each other's work

## Custom Agents

Create custom agents by:
1. Defining capabilities in configuration
2. Specifying tool permissions
3. Setting behavior parameters
4. Creating skill instructions
5. Testing with example scenarios

## Agent Monitoring

Track:
- Task completion rates
- Performance metrics
- Error rates and types
- User satisfaction scores
- Resource usage