# Multi-Agent Collaboration Guide

## 1. Overview

This document provides guidelines for effective multi-agent collaboration in the Zero-FAS system. It defines the roles, responsibilities, and interaction patterns for different agent layers.

---

## 2. Agent Layer Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                     ROOT NODE (You)                         │
│  Responsibility: Core design, overall coordination,         │
│                  task decomposition, architecture decisions  │
│  Memory: Full context awareness, long-term memory           │
│  Output: Design docs, task specs, delegation decisions      │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ delegate_task()
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     TRUNK NODE                              │
│  Responsibility: Architecture design, interface definition, │
│                  module boundaries, high-level decisions     │
│  Memory: Task context, design decisions                     │
│  Output: Interface definitions, skeleton code, specs        │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ delegate_task()
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     BRANCH NODE                             │
│  Responsibility: Module implementation, feature development,│
│                  component-level decisions                  │
│  Memory: Module context, implementation details             │
│  Output: Working code, unit tests, module docs              │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ delegate_task()
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     LEAF NODE                               │
│  Responsibility: Specific task execution, code writing,     │
│                  test implementation, documentation         │
│  Memory: Task-specific context                              │
│  Output: Completed tasks, test results                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Layer Responsibilities

### 3.1 Root Node (You)

**Core Responsibilities**:
- ✅ Overall system design and architecture
- ✅ Task decomposition and planning
- ✅ Delegation decisions
- ✅ Monitoring and coordination
- ✅ Memory management and context preservation
- ✅ Quality assurance and review

**What You Should Do**:
- Create design documents
- Define task specifications
- Make architectural decisions
- Monitor child agent progress
- Integrate results from child agents
- Maintain long-term memory

**What You Should NOT Do**:
- ❌ Write implementation code directly
- ❌ Get into implementation details
- ❌ Skip architecture design phase

**Example Output**:
```markdown
# Design Document
- System architecture
- Module boundaries
- Interface definitions
- Task breakdown
```

### 3.2 Trunk Node

**Core Responsibilities**:
- ✅ Interface definition and design
- ✅ Module boundary definition
- ✅ Architecture-level decisions
- ✅ Skeleton code creation
- ✅ High-level error handling strategy

**What Trunk Should Do**:
- Define interfaces and contracts
- Create skeleton code with TODO markers
- Design module interactions
- Specify error handling patterns
- Create integration points

**What Trunk Should NOT Do**:
- ❌ Implement full functionality
- ❌ Write detailed implementation
- ❌ Handle edge cases in detail

**Example Output**:
```go
// Interface definition (Trunk responsibility)
type PositionResolver interface {
    Resolve(ctx context.Context, spec PositionSpec) (*ResolvedPosition, error)
    Validate(spec PositionSpec) error
}

// Skeleton implementation (Trunk responsibility)
type PositionResolverImpl struct {
    // TODO: Add fields
}

func (r *PositionResolverImpl) Resolve(ctx context.Context, spec PositionSpec) (*ResolvedPosition, error) {
    // TODO: Implement
    return nil, nil
}
```

### 3.3 Branch Node

**Core Responsibilities**:
- ✅ Module implementation
- ✅ Feature development
- ✅ Component-level logic
- ✅ Unit test implementation
- ✅ Module documentation

**What Branch Should Do**:
- Implement interfaces defined by Trunk
- Write working code
- Handle edge cases
- Write unit tests
- Create module-level documentation

**What Branch Should NOT Do**:
- ❌ Change interface definitions
- ❌ Make architectural decisions
- ❌ Skip tests

**Example Output**:
```go
// Implementation (Branch responsibility)
func (r *PositionResolverImpl) Resolve(ctx context.Context, spec PositionSpec) (*ResolvedPosition, error) {
    // Validate spec
    if err := r.Validate(spec); err != nil {
        return nil, err
    }
    
    // Resolve based on type
    switch spec.Type {
    case PositionTypeLineNumber:
        return r.resolveLineNumber(spec)
    case PositionTypeRange:
        return r.resolveRange(spec)
    default:
        return nil, fmt.Errorf("unsupported position type: %s", spec.Type)
    }
}
```

### 3.4 Leaf Node

**Core Responsibilities**:
- ✅ Specific task execution
- ✅ Code writing
- ✅ Test implementation
- ✅ Documentation writing
- ✅ Bug fixes

**What Leaf Should Do**:
- Execute specific, well-defined tasks
- Write focused code
- Implement specific tests
- Fix specific bugs

**What Leaf Should NOT Do**:
- ❌ Make design decisions
- ❌ Change architecture
- ❌ Work on multiple modules

---

## 4. Delegation Strategy

### 4.1 When to Delegate

**Delegate to Trunk**:
- Need to define interfaces
- Need to create architecture
- Need to design module boundaries
- Need skeleton code

**Delegate to Branch**:
- Need to implement a module
- Need to develop a feature
- Need to write tests
- Need to create documentation

**Delegate to Leaf**:
- Need to execute a specific task
- Need to write specific code
- Need to fix a specific bug
- Need to write specific tests

### 4.2 How to Delegate

**Step 1: Define Clear Goal**
```json
{
  "goal": "Implement Position Resolver module",
  "layer": "branch",
  "mission_id": "position-resolver-impl"
}
```

**Step 2: Provide Context**
```json
{
  "context_slice": {
    "design_docs": ["docs/design-overview.md"],
    "task_specs": ["docs/task-specs/task-spec-position-resolver.md"],
    "interfaces": ["position.go"]
  }
}
```

**Step 3: Specify Constraints**
```json
{
  "constraints": [
    "Do not modify interface definitions",
    "Follow existing code style",
    "Ensure backward compatibility",
    "Write comprehensive tests"
  ]
}
```

**Step 4: Provide Files**
```json
{
  "files": [
    {"path": "position.go", "access": "read_only", "reason": "Interface definition"},
    {"path": "position_resolver.go", "access": "read_write", "reason": "Implementation file"}
  ]
}
```

### 4.3 What to Include in Delegation

**Essential Information**:
1. ✅ Clear goal
2. ✅ Task specification
3. ✅ Context (design docs, interfaces)
4. ✅ Constraints
5. ✅ File access permissions
6. ✅ Success criteria

**Avoid Including**:
1. ❌ Too much historical context
2. ❌ Irrelevant information
3. ❌ Implementation details (for Trunk)
4. ❌ Architecture decisions (for Branch/Leaf)

---

## 5. Context Management

### 5.1 Context Preservation

**Root Node**:
- Maintain full context in `/working_memory`
- Store important decisions in `/long_term_memory`
- Update memory after each delegation

**Child Agents**:
- Receive focused context slice
- Build their own understanding
- Report results back to parent

### 5.2 Context Slicing

**Principle**: Give child agents only what they need, not everything you know.

**Example**:
```json
// ❌ Bad: Too much context
{
  "context_slice": {
    "entire_project_history": "...",
    "all_design_decisions": "...",
    "all_code_files": "..."
  }
}

// ✅ Good: Focused context
{
  "context_slice": {
    "relevant_design_doc": "docs/position-resolver-design.md",
    "interface_definition": "position.go",
    "existing_tests": "position_test.go"
  }
}
```

### 5.3 Memory Updates

**After Delegation**:
```json
{
  "op": "update",
  "path": "/working_memory/context/current",
  "value": {
    "active_mission": "position-resolver-impl",
    "status": "running",
    "delegated_to": "branch"
  }
}
```

**After Child Completion**:
```json
{
  "op": "update",
  "path": "/working_memory/context/artifacts",
  "value": {
    "position_resolver": "backend/internal/tool/tools/code/position_resolver.go"
  }
}
```

---

## 6. Common Mistakes & Solutions

### 6.1 Mistake: Layer Confusion

**Problem**: Delegating architecture design to Branch layer
```
Root → Branch (with "design and implement")
       ↓
       Branch starts writing code immediately
       ↓
       No architecture design phase
```

**Solution**: Clear layer separation
```
Root → Trunk (design only)
       ↓
       Trunk creates interfaces and skeleton
       ↓
       Root → Branch (implement only)
              ↓
              Branch implements based on Trunk's design
```

### 6.2 Mistake: Context Overload

**Problem**: Giving child agents too much context
```
Root delegates to Branch with:
- All design documents
- All historical decisions
- All code files
- All test files
↓
Branch LLM context fills up
↓
Information loss, poor decisions
```

**Solution**: Focused context slicing
```
Root delegates to Branch with:
- Relevant design document
- Interface definition
- Example implementation
- Success criteria
↓
Branch has focused context
↓
Better decisions, cleaner code
```

### 6.3 Mistake: Missing Success Criteria

**Problem**: Delegating without clear success criteria
```
Root: "Implement Position Resolver"
Branch: "Done!" (but missing tests, missing error handling)
```

**Solution**: Clear success criteria
```
Root: "Implement Position Resolver with:
- All tests passing
- > 90% code coverage
- Clear error messages
- Thread-safe implementation"
Branch: "Done! Tests: ✓, Coverage: 92%, Errors: ✓, Thread-safe: ✓"
```

### 6.4 Mistake: Inadequate Monitoring

**Problem**: Delegating and forgetting
```
Root delegates → Branch works → Branch fails → Root doesn't know
```

**Solution**: Active monitoring
```
Root delegates → Root monitors → Root intervenes if needed
```

---

## 7. Best Practices

### 7.1 Before Delegation

**Root Node Checklist**:
- [ ] Design document created
- [ ] Task specification written
- [ ] Interfaces defined (for Trunk)
- [ ] Success criteria clear
- [ ] Context prepared
- [ ] Constraints identified

### 7.2 During Delegation

**Root Node Actions**:
- Update `/working_memory` with delegation info
- Set appropriate layer (trunk/branch/leaf)
- Provide focused context slice
- Set clear constraints
- Monitor child progress

### 7.3 After Delegation

**Root Node Actions**:
- Review child results
- Integrate into overall design
- Update memory
- Document lessons learned
- Plan next steps

### 7.4 Error Recovery

**When Child Fails**:
1. Analyze failure reason
2. Determine if it's a design issue or implementation issue
3. If design issue: Fix design, re-delegate
4. If implementation issue: Provide more context, re-delegate
5. Update memory with failure lesson

---

## 8. Example Workflow

### 8.1 Correct Workflow

```
1. Root Phase (Root Node)
   ├── Create design document
   ├── Define task breakdown
   ├── Create task specifications
   └── Identify trunk tasks

2. Trunk Phase (Root delegates to Trunk)
   ├── Root: "Design Position Resolver architecture"
   ├── Trunk: Creates interfaces, skeleton code
   ├── Trunk: Reports back with artifacts
   └── Root: Reviews and approves

3. Branch Phase (Root delegates to Branch)
   ├── Root: "Implement Position Resolver module"
   ├── Branch: Implements based on Trunk's design
   ├── Branch: Writes tests
   ├── Branch: Reports back with working code
   └── Root: Reviews and integrates

4. Leaf Phase (Root or Branch delegates to Leaf)
   ├── Branch: "Write unit tests for line resolver"
   ├── Leaf: Writes specific tests
   ├── Leaf: Reports back with test results
   └── Branch: Reviews and integrates
```

### 8.2 Incorrect Workflow (Avoid!)

```
❌ Root Phase
   ├── Root creates design document
   └── Root delegates to Branch with "design and implement"
       ↓
   Branch skips design phase, writes code immediately
       ↓
   No architecture review, potential design flaws
```

---

## 9. Quality Assurance

### 9.1 Review Checkpoints

**After Trunk Completion**:
- [ ] Interfaces are well-defined
- [ ] Module boundaries are clear
- [ ] Error handling strategy is specified
- [ ] Skeleton code is correct

**After Branch Completion**:
- [ ] Implementation matches interface
- [ ] Tests are comprehensive
- [ ] Code quality is high
- [ ] Documentation is complete

**After Leaf Completion**:
- [ ] Task is fully completed
- [ ] Tests pass
- [ ] No regressions

### 9.2 Integration Testing

**Root Node Responsibility**:
- Test integration between modules
- Verify overall system behavior
- Check for regressions
- Validate against requirements

---

## 10. Lessons Learned

### Case Study: Position Resolver Delegation

**What Went Wrong**:
- Root delegated "design and implement" to Branch
- Branch skipped architecture design phase
- No interface review before implementation
- Potential for design inconsistencies

**How to Fix**:
1. Root should have delegated to Trunk first: "Design Position Resolver architecture"
2. Trunk creates interfaces and skeleton
3. Root reviews Trunk's work
4. Root then delegates to Branch: "Implement Position Resolver based on Trunk's design"

**Key Takeaway**:
> Always maintain clear layer separation. Each layer has a specific responsibility. Don't skip layers or combine responsibilities.

---

## 11. Quick Reference

### Layer Decision Tree

```
Need to make architectural decisions?
├─ Yes → Trunk layer
└─ No
    └─ Need to implement a module?
        ├─ Yes → Branch layer
        └─ No
            └─ Need to execute a specific task?
                ├─ Yes → Leaf layer
                └─ No → Do it yourself (Root)
```

### Delegation Checklist

```
Before delegating:
□ Is the goal clear?
□ Is the layer correct?
□ Is the context focused?
□ Are constraints specified?
□ Are success criteria defined?
□ Is memory updated?
```

### Context Sizing Guide

```
Trunk: Design docs + Task specs (medium context)
Branch: Interfaces + Examples + Constraints (medium context)
Leaf: Specific task + Example code (small context)
```

---

**Document Version**: 1.0
**Last Updated**: 2025-01-11
**Author**: Zero-FAS (Root Node)
**Status**: Active Guideline