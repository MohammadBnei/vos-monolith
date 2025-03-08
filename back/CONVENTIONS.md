# Go Code Guidelines and Clean Architecture Directives

## Table of Contents

1. [General Principles](#general-principles)
2. [Clean Architecture Rules](#clean-architecture-rules)
3. [Uber Go Guidelines](#uber-go-guidelines)
4. [Testing Strategy](#testing-strategy)
5. [Documentation Standards](#documentation-standards)

## General Principles

### Clean Architecture Layering

- Core business logic at the center
- Interfaces define boundaries between layers
- Frameworks and drivers on outer layer
- Data flows inward, dependencies outward

### Package Organization

- Separate domains into distinct packages
- One public interface per package
- Internal packages for private implementation
- Clear separation of concerns

## Clean Architecture Rules

### Entities

- Represent core business domain
- Contain essential business rules
- Independent of infrastructure
- Immutable when possible

### Use Cases

- Define actions on entities
- Contain business rules
- Interface with repositories
- Handle input validation

### Gateways

- Abstract external services
- Implement interfaces defined by use cases
- Hide infrastructure details
- Manage data transformation

### Frameworks & Drivers

- External libraries and frameworks
- Database implementations
- Web frameworks
- File systems

## Uber Go Guidelines

### Error Handling

- Always handle errors explicitly
- Wrap errors with context using %w
- Log errors at appropriate level
- Return errors early

### Performance

- Avoid unnecessary allocations
- Use sync.Pool for frequent allocations
- Minimize goroutine creation
- Profile regularly

### Testing

- Write tests before implementation
- Mock interfaces at boundaries
- Test error paths thoroughly
- Maintain high coverage (>80%)

### Code Organization

- Keep functions short (<50 lines)
- Limit struct fields (<10)
- One concern per package
- Clear naming conventions

## Testing Strategy

### Unit Tests

- Focus on individual components
- Use mocks for dependencies
- Test both success and failure paths
- Keep tests fast and isolated

### Integration Tests

- Test across layer boundaries
- Verify component interactions
- Validate error propagation
- Cover critical business flows

### End-to-End Tests

- Test complete user journeys
- Verify system behavior
- Run periodically
- Monitor performance metrics

## Documentation Standards

### Public Interfaces

- Document all exported types
- Explain method parameters
- Describe return values
- Include usage examples

### Private Implementation

- Comment complex algorithms
- Explain design decisions
- Document performance considerations
- Note potential pitfalls

## Commit Messages

- Follow conventional commits spec
- Include type prefix (feat, fix, docs, etc.)
- Brief summary under 50 chars
- Detailed body after blank line
