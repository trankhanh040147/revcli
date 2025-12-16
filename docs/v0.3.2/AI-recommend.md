
### Analysis of Proposed Features

#### 1. Prompt First (Optional)

- **Why:** Currently, your CLI likely dumps a general review. Users often want specific things (e.g., "Check for thread safety" or "Is this strictly following Clean Architecture?").
- **UX Improvement:** Instead of a simple text input, offer a **"Pre-flight Checklist"**. Allow the user to type a custom prompt OR select quick tags (e.g., `[ ] Security`, `[ ] Performance`, `[ ] Naming`).

#### 2. Reviews Interaction (The "Ignore" Logic)

- **The Problem:** "Ignore from context" implies you need to modify the context window dynamically _during_ a session.
    
- **Refinement - "Context Pruning":**
    
    - **Ignore from Context:** Don't just delete it. Replace the huge code block with a one-line summary (AI-generated) to save tokens while keeping the logic flow.
        
    - **Ignore from System/Preset:** This is essentially "Negative Prompting." If I say "Ignore naming conventions," you append "Do not report naming convention errors" to the system prompt.
        

### New Tech Recommendations

Since you love new tech, here are libraries to supercharge v0.3.1:

1. **`charmbracelet/huh`** (Form/Input)
    
    - **Use for:** The "Prompt First" feature.
        
    - **Why:** It’s a newer library from Charm (Bubbletea creators). It builds accessible, beautiful forms (selects, text inputs, confirms) with zero boilerplate. It fits your "Keyboard-First" UX perfectly.
        
    - **UX Win:** A clean, modal-like form pops up before the heavy AI lifting starts.
        
2. **`samber/lo`** (Utility)
    
    - **Use for:** Your "Refactor" phase.
        
    - **Why:** It's a Lodash-style library for Go using Generics. It makes slice/map manipulation (filtering reviews, mapping ignored files) incredibly concise and readable.
        
    - **Example:** `lo.Filter(reviews, func(x, _ int) bool { return !x.Ignored })`
        
3. **`dgraph-io/badger`** or **`nutsdb`** (Embedded DB)
    
    - **Use for:** "Prompt Memory."
        
    - **Why:** To make "Ignore from preset/system" persistent, you need state. A pure YAML file is okay, but an embedded KV store allows you to cache "context summaries" (see below) so you don't re-summarize files you've already seen.
        

---

### Revised Roadmap: v0.3.1 - Context & Intent

I renamed it to "Context & Intent" because "Memory" implies vector DBs, whereas here we are managing **what** we send to the AI.

**Status:** Refining

**Features:**

#### 🎯 Intent-Driven Review (New "Prompt First")

- [ ] **Pre-Review Form (`huh`):** Before scanning, ask:
    - Custom instruction (e.g., "Focus on error handling").
    - Select Focus Areas (Security, Perf, Logic).
- [ ] **Smart Context:** If the user asks for "Security," automatically inject the `security` preset rules into the system prompt.

#### 🧠 Context Pruning (Dynamic Ignore)

- [ ] **"Summarize & Prune" Action:** In the TUI, pressing `i` on a file/block:
    1. Uses a cheap model (Gemini Flash) to summarize the code block (e.g., `// func processData: handles standard CSV parsing`).
    2. Replaces the actual code in the context window with this summary.
    3. **Benefit:** Saves massive tokens for the _next_ turn of chat while keeping the "map" of the code.
- [ ] **Negative Prompting:** "Ignore from System" adds a negative constraint to the session config (e.g., "User explicitly stated they don't care about variable names").

