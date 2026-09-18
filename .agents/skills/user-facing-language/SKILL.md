---
name: user-facing-language
description: Write or review user-facing product copy (for example labels, descriptions, status, errors, alerts). Use for UI text, not internal IDs, APIs, or dev docs.
---

# User-facing language

Write for the person using the product, not for its implementation. Prefer familiar user and application terms, concrete actions, and visible consequences over internal, abstract, or technical vocabulary.

## Choose clear, user-centered terms

- Name what users recognize and can act on. Prefer **Team member** over **User entity**, and **unsaved changes** over **working copy**.
- State the useful outcome or restriction directly. Prefer **You need admin permissions to edit this page** to an explanation of which backend authorization policy failed.
- Explain domain rules by user impact. For example, say **Pausing a subscription keeps your access active until the billing cycle ends**, rather than explaining how the cancellation flag operates independently from access entitlement.
- Use terminology consistently across all views unless a more specific established term is needed.

## Treat descriptions as optional

Do not add a description merely because a component supports one. Omit it when the title, control label, surrounding context, or immediate result already makes the purpose clear.

Add a description only when it helps a user make a decision or understand one of these:

- What the action affects
- A consequence, limitation, or requirement that is not otherwise visible
- A non-obvious distinction between choices.

Avoid descriptions that simply restate their title or control, add generic filler, or narrate obvious UI mechanics. When a description is warranted, make it concise and describe the user-relevant outcome.

## Review before handoff

For new or revised UI copy, scan the changed labels, helper text, empty states, toasts, confirmations, errors, and accessible names:

- Replace internal jargon with clear user concepts.
- Remove redundant descriptions.
- Verify accuracy and consistent use of terminology.